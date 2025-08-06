package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "gowok",
	Short: "Gowok is a Go project bootstrapper",
	Long:  `A fast and flexible CLI for Go projects`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "There was an error while executing gowok: %s", err)
		os.Exit(1)
	}
}
