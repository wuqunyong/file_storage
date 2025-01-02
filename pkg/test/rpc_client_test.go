package test

import (
	"fmt"
	"testing"
	"time"

	"github.com/wuqunyong/file_storage/pkg/actor"
	"github.com/wuqunyong/file_storage/pkg/rpc"
)

func TestClient1(t *testing.T) {
	//http://127.0.0.1:8222/connz?subs=true
	sTopic := "identify.server.1.2.3"

	engine := actor.NewEngine("test", actor.LocalLookupAddr, "")
	rpcClient := rpc.NewRPCClient(engine, "nats://127.0.0.1:4222", "identify.client.1.2.3")
	err := rpcClient.Init()
	if err != nil {
		t.Fatalf("err:%s", err)
	}
	go rpcClient.Run()

	time.Sleep(time.Duration(3) * time.Second)

	rpcClient.Send(sTopic, []byte("test client 1"))
	rpcClient.Send(sTopic, []byte("test client 2"))
	rpcClient.Send(sTopic, []byte("test client 3"))

	time.Sleep(time.Duration(180) * time.Second)
}

func TestClient2(t *testing.T) {
	//http://127.0.0.1:8222/connz?subs=true
	engine := actor.NewEngine("test", actor.LocalLookupAddr, "")
	rpcClient := rpc.NewRPCClient(engine, "nats://127.0.0.1:4222", "identify.client.1.2.3")
	err := rpcClient.Init()
	if err != nil {
		t.Fatalf("err:%s", err)
	}
	go rpcClient.Run()

	time.Sleep(time.Duration(10) * time.Second)

	rpcClient.Stop()

	fmt.Printf("rpc closed")
	time.Sleep(time.Duration(6) * time.Second)
}
