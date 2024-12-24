/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"

	"github.com/spf13/cobra"
	"github.com/wuqunyong/file_storage/pkg/actor"
	"github.com/wuqunyong/file_storage/pkg/common"
	"github.com/wuqunyong/file_storage/pkg/component/mongodb"
	"github.com/wuqunyong/file_storage/pkg/ws"
)

// apiCmd represents the api command
var apiCmd = &cobra.Command{
	Use:   "api",
	Short: "A brief description of your command",
	Long:  `A longer description that spans multiple lines`,
	Run: func(cmd *cobra.Command, args []string) {
		var config ws.Config
		// 使用 Flags
		config.HttpPort, _ = cmd.Flags().GetString("httpPort")
		config.HttpsPort, _ = cmd.Flags().GetString("HttpsPort")
		config.ServerCertificate = "E:/VCity/city/config/metaserver.vcity.app_chain.crt"
		config.ServerPrivateKey = "E:/VCity/city/config/metaserver.vcity.app_key.key"

		engine := actor.NewEngine("test", "1.2.3", "nats://127.0.0.1:4222")

		var mongoConfig mongodb.Config
		mongoConfig.Uri = "mongodb://admin:123456@127.0.0.1:27018"
		mongoConfig.Database = "vcity"
		component := mongodb.NewMongoComponent(context.Background(), &mongoConfig)
		engine.AddComponent(component)

		err := engine.Init()
		if err != nil {
			return
		}
		engine.Start()
		defer engine.Stop()

		wsServer := ws.NewWsServer(config, engine)
		wsServer.Run()
		defer wsServer.Stop()

		// moduleA := &ws.ModuleA{}
		// ws.GetInstance().Register(1, moduleA.Handler_Func1)
		// ws.GetInstance().Register(2, moduleA.Handler_Func2)
		common.WaitForShutdown()
	},
}

func init() {
	// 添加子命令的 Flags
	apiCmd.Flags().StringP("httpPort", "", ":8080", "httpPort")
	apiCmd.Flags().StringP("HttpsPort", "", ":8081", "HttpsPort")

	rootCmd.AddCommand(apiCmd)
}
