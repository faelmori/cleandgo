package cli

import (
	"github.com/spf13/cobra"
)

func ServiceCmdList() []*cobra.Command {
	return []*cobra.Command{
		startCommand(),
	}
}

func startCommand() *cobra.Command {
	var name, port, bind, logFile, configFile string
	var isConfidential, debug bool

	var startCmd = &cobra.Command{
		Use: "start",
		Annotations: GetDescriptions([]string{
			"Start a minimal backend service",
			"Start a minimal backend service with CleandGO",
		}, false),
		Run: func(cmd *cobra.Command, args []string) {

		},
	}

	startCmd.Flags().StringVarP(&name, "name", "n", "CleandGO", "Name of the process")
	startCmd.Flags().StringVarP(&port, "port", "p", ":8666", "Port to listen on")
	startCmd.Flags().StringVarP(&bind, "bind", "b", "0.0.0.0", "Bind address")
	startCmd.Flags().StringVarP(&logFile, "log-file", "l", "", "Log file path")
	startCmd.Flags().StringVarP(&configFile, "config-file", "c", "", "Configuration file path")
	startCmd.Flags().BoolVarP(&isConfidential, "confidential", "C", false, "Enable confidential mode")
	startCmd.Flags().BoolVarP(&debug, "debug", "d", false, "Enable debug mode")

	return startCmd
}
