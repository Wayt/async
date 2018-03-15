package cmd

import (
	"github.com/spf13/cobra"
	"github.com/wayt/async/server"
)

func init() {
	rootCmd.AddCommand(serverCmd)
}

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Runs Async server daemon",
	Run: func(cmd *cobra.Command, args []string) {

		s := server.New()
		s.Run()
	},
}
