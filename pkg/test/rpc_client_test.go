package test

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"testing"
	"time"

	"github.com/wuqunyong/file_storage/pkg/actor"
	"github.com/wuqunyong/file_storage/pkg/concepts"
	"github.com/wuqunyong/file_storage/pkg/rpc"
	"github.com/wuqunyong/file_storage/proto/rpc_msg"

	"github.com/wuqunyong/file_storage/proto/common_msg"
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

	// person := &common_msg.Person{Name: "小明", Age: 123456}
	// obj1, err := actor.SendRequest[common_msg.Person](actorObj1, concepts.NewActorId("engine.0.1.1001.server", "1"), 1, person)
	// if err != nil {
	// 	fmt.Printf("\n\n\n================obj1:err:%v,data:%v\n", err, obj1)
	// } else {
	// 	fmt.Printf("\n\n\n================obj1:%T, %v\n", obj1, obj1)
	// }

	// obj2, err := actor.SendRequest[common_msg.Person](actorObj1, concepts.NewActorId("engine.0.1.1001.server", "1"), 2, person)
	// if err != nil {
	// 	fmt.Printf("\n\n\n================obj2:err:%v,data:%v\n", err, obj2)
	// } else {
	// 	fmt.Printf("\n\n\n================obj2:%T, %v\n", obj2, obj2)
	// }

	echoObj := &rpc_msg.RPC_EchoTestRequest{Value1: 12345678, Value2: "小明"}

	//engine.1.4.1.serve
	echoResponse, err := actor.SendRequest[rpc_msg.RPC_EchoTestResponse](actorObj1, concepts.NewActorId("engine.0.1.1001.server", "1"), 1001, echoObj)
	// echoResponse, err := actor.SendRequest[rpc_msg.RPC_EchoTestResponse](actorObj1, concepts.NewActorId("engine.1.4.1.server", "C++"), 1001, echoObj)
	if err != nil {
		fmt.Println("err", err)
	}
	fmt.Printf("\n\n\n================echoResponse:%T, %v\n", echoResponse, echoResponse)

	sendErr := actor.SendNotify(actorObj1, concepts.NewActorId("engine.0.1.1001.server", "1"), 1002, echoObj)
	if sendErr != nil {
		fmt.Printf("\n\n\n================sendNotify Error:%T, %v\n", echoObj, echoObj)
	} else {
		fmt.Printf("\n\n\n================sendNotify Success:%T, %v\n", echoObj, echoObj)
	}

	time.Sleep(time.Duration(1800) * time.Second)
	time.Sleep(time.Duration(1800) * time.Second)
}

func TestClientRegister(t *testing.T) {
	engine := actor.NewEngine(0, 1, 1002, "nats://127.0.0.1:4222")

	engine.MustInit()

	actorObj1 := &ActorObjA{
		Actor: actor.NewActor("1", engine),
	}
	engine.SpawnActor(actorObj1)

	engine.Start()

	nodeObj := &common_msg.MSG_REQUEST_REGISTER_INSTANCE{
		Instance: &common_msg.EndPointInstance{Realm: 1, Type: 4, Id: 1},
		Auth:     "hello",
	}

	//engine.1.1.1.serve
	nodeResponse, err := actor.SendRequest[common_msg.MSG_RESPONSE_REGISTER_INSTANCE](actorObj1, concepts.NewActorId("engine.1.1.1.server", "C++"), 410, nodeObj)
	if err != nil {
		fmt.Println("err", err)
	}
	fmt.Printf("\n\n\n================nodeResponse:%T, %v\n", nodeResponse, nodeResponse)

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs
}
