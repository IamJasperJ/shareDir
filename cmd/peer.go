package cmd

import (
	"fmt"

	"shareDir/daemon"

	"github.com/spf13/cobra"
)

var (
	peerID      string
	trackerAddr string
	peerDaemon  bool
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

		behavior := daemon.NewPeerBehavior(peerID, trackerAddr)
		d := daemon.New("peer", behavior)

		return d.Run(peerDaemon)
	},
}

var peerStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the peer",
	RunE: func(cmd *cobra.Command, args []string) error {
		d := daemon.New("peer", nil)
		return d.Stop()
	},
}

func init() {
	peerStartCmd.Flags().StringVar(&peerID, "id", "", "Peer ID")
	peerStartCmd.Flags().StringVar(&trackerAddr, "addr", "", "Tracker address (e.g. 127.0.0.1:8080)")
	peerStartCmd.Flags().BoolVarP(&peerDaemon, "daemon", "d", false, "Run in background")

	peerCmd.AddCommand(peerStartCmd, peerStopCmd)
	RootCmd.AddCommand(peerCmd)
}
