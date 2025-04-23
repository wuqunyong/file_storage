package test

import (
	"errors"
	"fmt"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"testing"
	"time"

	"github.com/wuqunyong/file_storage/pkg/actor"
	"github.com/wuqunyong/file_storage/pkg/concepts"
	logger "github.com/wuqunyong/file_storage/pkg/logger"
	"github.com/wuqunyong/file_storage/proto/rpc_msg"

	"github.com/wuqunyong/file_storage/proto/common_msg"
)

type ActorClient struct {
	*actor.Actor
	inited atomic.Bool
}

func (actor *ActorClient) OnInit() error {
	if actor.inited.Load() {
		return errors.New("duplicate init")
	}
	actor.inited.Store(true)
	return nil
}

func (actor *ActorClient) OnShutdown() {

}

func TestClient1(t *testing.T) {
	engine := actor.NewEngine(0, 1, 1002, "nats://127.0.0.1:4222")

	engine.MustInit()
	defer engine.Stop()
	engine.Start()

	actorObj1 := &ActorClient{
		Actor: actor.NewActor("1", engine),
	}
	//engine.SpawnActor(actorObj1)

	echo := &common_msg.EchoRequest{Value1: 123456, Value2: "小明"}
	obj1, err := actor.SendRequest[common_msg.EchoResponse](actorObj1, concepts.NewActorId("engine.0.1.1001.server", "1"), 1, echo)
	if err != nil {
		logger.Log(logger.InfoLevel, "opcode 1 failure", "request", echo, "response", obj1, "err", err)
	} else {
		logger.Log(logger.InfoLevel, "opcode 1 success", "request", echo, "response", obj1)
	}

	obj2, err := actor.SendRequest[common_msg.EchoResponse](actorObj1, concepts.NewActorId("engine.0.1.1001.server", "1"), 2, echo)
	if err != nil {
		logger.Log(logger.InfoLevel, "opcode 2 failure", "request", echo, "response", obj2, "err", err)
	} else {
		logger.Log(logger.InfoLevel, "opcode 2 success", "request", echo, "response", obj2)
	}

	echoObj := &rpc_msg.RPC_EchoTestRequest{Value1: 12345678, Value2: "小明"}
	//engine.1.4.1.serve
	echoResponse, err := actor.SendRequest[rpc_msg.RPC_EchoTestResponse](actorObj1, concepts.NewActorId("engine.0.1.1001.server", "1"), 1001, echoObj)
	// echoResponse, err := actor.SendRequest[rpc_msg.RPC_EchoTestResponse](actorObj1, concepts.NewActorId("engine.1.4.1.server", "C++"), 1001, echoObj)
	if err != nil {
		logger.Log(logger.InfoLevel, "opcode 1001 failure", "request", echoObj, "response", echoResponse, "err", err)
	} else {
		logger.Log(logger.InfoLevel, "opcode 1001 success", "request", echoObj, "response", echoResponse)
	}

	sendErr := actor.SendNotify(actorObj1, concepts.NewActorId("engine.0.1.1001.server", "1"), 1002, echoObj)
	if sendErr != nil {
		logger.Log(logger.InfoLevel, "opcode 1002 failure", "notify", echoObj, "err", sendErr)
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
