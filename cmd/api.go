/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/wuqunyong/file_storage/pkg/ws"
)

// apiCmd represents the api command
var apiCmd = &cobra.Command{
	Use:   "api",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		var config ws.Config
		// 使用 Flags
		config.HttpPort, _ = cmd.Flags().GetString("httpPort")
		config.HttpsPort, _ = cmd.Flags().GetString("HttpsPort")
		config.ServerCertificate = "E:/VCity/city/config/metaserver.vcity.app_chain.crt"
		config.ServerPrivateKey = "E:/VCity/city/config/metaserver.vcity.app_key.key"
		wsServer := ws.NewWsServer(config)
		wsServer.Run()
		defer wsServer.Stop()

		moduleA := &ws.ModuleA{}
		ws.GetInstance().Register(1, moduleA.Handler_Func1)
		ws.GetInstance().Register(2, moduleA.Handler_Func2)

		// Wait for the process to be shutdown.
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
		<-sigs
		ws.SIGTERMExit()
	},
}

func init() {
	// 添加子命令的 Flags
	apiCmd.Flags().StringP("httpPort", "", ":8080", "httpPort")
	apiCmd.Flags().StringP("HttpsPort", "", ":8081", "HttpsPort")

	rootCmd.AddCommand(apiCmd)
}
