package cmd

import "github.com/spf13/cobra"

var RootCmd = &cobra.Command{
	Use:   "share",
	Short: "p2p method to share directories.",
}
