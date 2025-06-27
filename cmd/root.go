package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// represent the root command of cli application
var rootCmd = &cobra.Command{
	Use:   "tfcount",
	Short: "A simple CLI to summarize terraform plan outputs by resource type and action",
	Run: func(cmd *cobra.Command, args []string) {
		// logic for root cmd
	},
}

// entrypoint for the CLI
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
