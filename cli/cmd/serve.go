package cmd

import (
	"fmt"

	"github.com/gowok/gowok"
	"github.com/spf13/cobra"
)

var configFile string
var envFile string

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Serve the gowok application",
	Long:  `Serve the gowok application by specifying the configuration file.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(configFile, envFile)
		gowok.Get().WithConfig(configFile, envFile).Run()
	},
}

func init() {
	serveCmd.Flags().StringVarP(&configFile, "config", "c", "", "Yaml configuration file location")
	serveCmd.Flags().StringVarP(&envFile, "env-file", "e", "", "Env file location")
	rootCmd.AddCommand(serveCmd)
}
