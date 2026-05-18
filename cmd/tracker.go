package cmd

import (
	"fmt"
	"shareDir/daemon"
	"strconv"

	"github.com/spf13/cobra"
)

var (
	port          string
	trackerDaemon bool
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

		behavior := daemon.NewTrackerBehavior(portnum)
		d := daemon.New("tracker", behavior)

		return d.Run(trackerDaemon)
	},
}

var trackerStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the tracker",
	RunE: func(cmd *cobra.Command, args []string) error {
		d := daemon.New("tracker", nil)
		return d.Stop()
	},
}

func init() {
	trackerStartCmd.Flags().StringVar(&port, "port", "", "Port for tracker to listen on")
	trackerStartCmd.Flags().BoolVarP(&trackerDaemon, "daemon", "d", false, "Run in background")

	trackerCmd.AddCommand(trackerStartCmd, trackerStopCmd)
	RootCmd.AddCommand(trackerCmd)
}
