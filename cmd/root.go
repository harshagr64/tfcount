package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// represent the root command of cli application
var showVersion bool
var version = "dev"
var rootCmd = &cobra.Command{
	Use:   "tfcount",
	Short: "A simple CLI to summarize terraform/terragrunt plan outputs by resource type and action",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if showVersion {
			fmt.Println(cmd.Name(), version)
			os.Exit(0)
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func init() {
	// add version flag to root command
	rootCmd.PersistentFlags().BoolVarP(&showVersion, "version", "v", false, "Show version information")
}

// entrypoint for the CLI
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
