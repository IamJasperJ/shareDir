package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"shareDir/iptracker/tracker"
	"strconv"
	"syscall"

	"github.com/spf13/cobra"
)

var (
	port string
)

var trackerCmd = &cobra.Command{
	Use:   "tracker",
	Short: "Run tracker",
}

var trackerStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the tracker",
	RunE: func(cmd *cobra.Command, args []string) error {
		if port == "" {
			return fmt.Errorf("--port is required")
		}

		portnum, err := strconv.Atoi(port)
		if err != nil || portnum > 65535 {
			return fmt.Errorf("port should be an available port number")
		}
		t := tracker.New(portnum)

		pid := os.Getpid()
		pidFile := "/tmp/share_tracker.pid"
		if err := os.WriteFile(pidFile, []byte(strconv.Itoa(pid)), 0644); err != nil {
			return fmt.Errorf("write pid file: %w", err)
		}
		defer os.Remove(pidFile)

		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

		go func() {
			sig := <-sigCh
			fmt.Println("\nReceived signal:", sig)
			t.Stop(fmt.Errorf("signal %v", sig))
		}()

		if err := t.Start(); err != nil {
			return err
		}
		return nil
	},
}

var trackerStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the tracker",
	RunE: func(cmd *cobra.Command, args []string) error {
		pidFile := "/tmp/share_tracker.pid"
		data, err := os.ReadFile(pidFile)
		if err != nil {
			return fmt.Errorf("tracker may not be running (pid file not found): %w", err)
		}

		pid, err := strconv.Atoi(string(data))
		if err != nil {
			return fmt.Errorf("invalid pid file: %w", err)
		}

		proc, err := os.FindProcess(pid)
		if err != nil {
			return fmt.Errorf("find process %d: %w", pid, err)
		}

		if err := proc.Signal(syscall.SIGTERM); err != nil {
			return fmt.Errorf("signal process %d: %w", pid, err)
		}

		os.Remove(pidFile)
		fmt.Println("Tracker stopped (pid", pid, ")")
		return nil
	},
}

func init() {
	trackerStartCmd.Flags().StringVar(&port, "port", "", "Port for tracker to listen on")

	trackerCmd.AddCommand(trackerStartCmd, trackerStopCmd)
	RootCmd.AddCommand(trackerCmd)
}
