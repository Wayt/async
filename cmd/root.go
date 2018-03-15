package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "async",
	Short: "Async is a framework that let you build distributed function pipelines. ",
	Long: `A fast and distributed pipeline framework that let you implement your business workload the way you want.
See https://github.com/wayt/async`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
		os.Exit(2)
	},
}

func Execute() {

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
