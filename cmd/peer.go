package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"shareDir/iptracker/peer"

	"github.com/spf13/cobra"
)

var (
	peerID    string
	trackerAddr string
)

var peerCmd = &cobra.Command{
	Use:   "peer",
	Short: "Run peer",
}

var peerStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the peer",
	RunE: func(cmd *cobra.Command, args []string) error {
		if peerID == "" || trackerAddr == "" {
			return fmt.Errorf("--id and --addr are required")
		}

		p := peer.New(peerID, trackerAddr)

		pid := os.Getpid()
		pidFile := "/tmp/share_peer.pid"
		if err := os.WriteFile(pidFile, []byte(strconv.Itoa(pid)), 0644); err != nil {
			return fmt.Errorf("write pid file: %w", err)
		}
		defer os.Remove(pidFile)

		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

		go func() {
			sig := <-sigCh
			fmt.Println("\nReceived signal:", sig)
			p.Stop(fmt.Errorf("signal %v", sig))
		}()

		p.Start()
		return nil
	},
}

var peerStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the peer",
	RunE: func(cmd *cobra.Command, args []string) error {
		pidFile := "/tmp/share_peer.pid"
		data, err := os.ReadFile(pidFile)
		if err != nil {
			return fmt.Errorf("peer may not be running (pid file not found): %w", err)
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
		fmt.Println("Peer stopped (pid", pid, ")")
		return nil
	},
}

func init() {
	peerStartCmd.Flags().StringVar(&peerID, "id", "", "Peer ID")
	peerStartCmd.Flags().StringVar(&trackerAddr, "addr", "", "Tracker address (e.g. 127.0.0.1:8080)")

	peerCmd.AddCommand(peerStartCmd, peerStopCmd)
	RootCmd.AddCommand(peerCmd)
}
