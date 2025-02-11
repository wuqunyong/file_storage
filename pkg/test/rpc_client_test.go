package test

import (
	"fmt"
	"testing"
	"time"

	"github.com/wuqunyong/file_storage/pkg/actor"
	"github.com/wuqunyong/file_storage/pkg/concepts"
	"github.com/wuqunyong/file_storage/pkg/rpc"
	"github.com/wuqunyong/file_storage/rpc_msg"
)

func TestClient1(t *testing.T) {
	//http://127.0.0.1:8222/connz?subs=true
	//sTopic := "identify.server.1.2.3"
	sTopic := concepts.GenServerAddress(0, 1, 1001)

	engine := actor.NewEngine(0, 1, 1001, "")
	sClientAddress := concepts.GenClientAddress(0, 1, 1001)
	rpcClient := rpc.NewRPCClient(engine, "nats://127.0.0.1:4222", sClientAddress)
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
	engine := actor.NewEngine(0, 1, 1001, "")
	sClientAddress := concepts.GenClientAddress(0, 1, 1001)
	rpcClient := rpc.NewRPCClient(engine, "nats://127.0.0.1:4222", sClientAddress)
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

func TestClient3(t *testing.T) {
	engine := actor.NewEngine(0, 1, 1002, "nats://127.0.0.1:4222")

	engine.MustInit()
	engine.Start()

	actorObj1 := &ActorObjA{
		Actor: actor.NewActor("1", engine),
	}
	//engine.SpawnActor(actorObj1)

	// person := &testdata.Person{Name: "小明", Age: 123456}
	// obj, err := actor.SendRequest[testdata.Person](actorObj1, concepts.NewActorId("engine.0.1.1001.server", "1"), 1, person)
	// if err != nil {
	// 	fmt.Println("err", err)
	// }
	// fmt.Printf("\n\n\n================obj:%T, %v\n", obj, obj)

	echoObj := &rpc_msg.RPC_EchoTestRequest{Value1: 12345678, Value2: "小明"}

	//engine.1.4.1.serve
	echoResponse, err := actor.SendRequest[rpc_msg.RPC_EchoTestResponse](actorObj1, concepts.NewActorId("engine.0.1.1001.server", "1"), 1001, echoObj)
	//echoResponse, err := actor.SendRequest[rpc_msg.RPC_EchoTestResponse](actorObj1, concepts.NewActorId("engine.1.4.1.server", "C++"), 1001, echoObj)
	if err != nil {
		fmt.Println("err", err)
	}
	fmt.Printf("\n\n\n================echoResponse:%T, %v\n", echoResponse, echoResponse)

	time.Sleep(time.Duration(1800) * time.Second)
	time.Sleep(time.Duration(1800) * time.Second)
}
