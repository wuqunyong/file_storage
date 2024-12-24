package rpc

import (
	"fmt"
	"testing"
	"time"

	"github.com/wuqunyong/file_storage/pkg/actor"
)

func TestServer1(t *testing.T) {
	//http://127.0.0.1:8222/connz?subs=true
	engine := actor.NewEngine(actor.LocalLookupAddr, "")
	rpcServer := NewRPCServer(engine, "nats://127.0.0.1:4222", "identify.server.1.2.3")
	err := rpcServer.Init()
	if err != nil {
		t.Fatalf("err:%s", err)
	}
	rpcServer.Run()
}

func TestServer2(t *testing.T) {
	//http://127.0.0.1:8222/connz?subs=true
	//nats pub identify.server.1.2.3 "hello world"
	engine := actor.NewEngine(actor.LocalLookupAddr, "")
	rpcServer := NewRPCServer(engine, "nats://127.0.0.1:4222", "identify.server.1.2.3")
	err := rpcServer.Init()
	if err != nil {
		t.Fatalf("err:%s", err)
	}
	go rpcServer.Run()

	time.Sleep(time.Duration(10) * time.Second)

	rpcServer.Stop()

	fmt.Printf("rpc closed")
	time.Sleep(time.Duration(6) * time.Second)
}
