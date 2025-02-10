/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"

	"github.com/spf13/cobra"
	"github.com/wuqunyong/file_storage/pkg/actor"
	"github.com/wuqunyong/file_storage/pkg/cluster/discovery/registry"
	"github.com/wuqunyong/file_storage/pkg/component/etcd"
	"github.com/wuqunyong/file_storage/pkg/component/mongodb"
	"github.com/wuqunyong/file_storage/pkg/component/tcpserver"
	"github.com/wuqunyong/file_storage/pkg/component/wsserver"
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
		config.HttpPort, _ = cmd.Flags().GetString("HttpPort")
		config.HttpsPort, _ = cmd.Flags().GetString("HttpsPort")
		config.ServerCertificate = "E:/VCity/city/config/metaserver.vcity.app_chain.crt"
		config.ServerPrivateKey = "E:/VCity/city/config/metaserver.vcity.app_key.key"

		engine := actor.NewEngine(0, 1, 1001, "nats://127.0.0.1:4222")

		configs := map[string]*mongodb.Config{}
		configs["test"] = &mongodb.Config{
			Uri:        "mongodb://admin:123456@127.0.0.1:27018",
			Database:   "vcity",
			ConnectNum: 2,
		}
		mongoComp := mongodb.NewMongoComponent(context.Background(), configs)
		engine.MustAddComponent(mongoComp)

		tcpServerComp := tcpserver.NewTCPServer(tcpserver.NewPBServerOption(), ":16007")
		engine.MustAddComponent(tcpServerComp)

		wsServerComp := wsserver.NewWSServer(config)
		engine.MustAddComponent(wsServerComp)

		var opts []registry.Option
		opts = append(opts, registry.Addrs("127.0.0.1:2379"))

		etcdComp := etcd.NewEtcvServiceDiscovery(opts...)
		engine.MustAddComponent(etcdComp)

		engine.MustInit()
		engine.Start()
		defer engine.Stop()

		engine.WaitForShutdown()
	},
}

func init() {
	// 添加子命令的 Flags
	apiCmd.Flags().StringP("HttpPort", "", ":8080", "HttpPort")
	apiCmd.Flags().StringP("HttpsPort", "", ":8081", "HttpsPort")

	rootCmd.AddCommand(apiCmd)
}
