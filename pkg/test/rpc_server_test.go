package test

import (
	"context"
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
	"github.com/wuqunyong/file_storage/pkg/errs"
	"github.com/wuqunyong/file_storage/pkg/logger"
	"github.com/wuqunyong/file_storage/pkg/rpc"
	"github.com/wuqunyong/file_storage/proto/common_msg"
	"github.com/wuqunyong/file_storage/proto/rpc_msg"
)

type ActorObjB struct {
	*actor.Actor
	inited atomic.Bool
}

func (actor *ActorObjB) OnInit() error {
	if actor.inited.Load() {
		return errors.New("duplicate init")
	}
	actor.Register(1, actor.Func1)
	actor.Register(2, actor.Func2)
	actor.Register(1001, actor.EchoTest)
	actor.Register(1002, actor.NotifyTest)
	actor.inited.Store(true)
	return nil
}

func (actor *ActorObjB) OnShutdown() {
	fmt.Printf("OnShutdown\n")
}

func (actor *ActorObjB) Func1(ctx context.Context, request *common_msg.EchoRequest, reply *common_msg.EchoResponse) errs.CodeError {
	reply.Value1 = request.Value1 + 1
	reply.Value2 = request.Value2 + " | response"
	fmt.Printf("reply value:%v\n", reply)

	return nil
}

func (actor *ActorObjB) Func2(ctx context.Context, request *common_msg.EchoRequest, reply *common_msg.EchoResponse) errs.CodeError {

	fmt.Printf("request value:%v\n", request)
	return errs.NewCodeError(errors.New("invalid"), 123)
}

func (actor *ActorObjB) EchoTest(ctx context.Context, arg *rpc_msg.RPC_EchoTestRequest, reply *rpc_msg.RPC_EchoTestResponse) errs.CodeError {
	reply.Value1 = arg.Value1
	reply.Value2 = arg.Value2 + "|Response"
	fmt.Printf("inside value:%v\n", reply)

	return nil
}

func (actor *ActorObjB) NotifyTest(ctx context.Context, arg *rpc_msg.RPC_EchoTestRequest) {
	fmt.Printf("notify value:%v\n", arg)
}

func (actor *ActorObjB) Func3() {
	fmt.Println("func3")
}

func TestServer1(t *testing.T) {
	//http://127.0.0.1:8222/connz?subs=true

	logger.Log(logger.DebugLevel, "TestServer1 Debug")
	logger.Log(logger.InfoLevel, "TestServer1 Info")
	engine := actor.NewEngine(0, 1, 1001, "nats://127.0.0.1:4222")
	engine.MustInit()
	actorObj1 := &ActorObjB{
		Actor: actor.NewActor("1", engine),
	}
	engine.SpawnActor(actorObj1)
	defer engine.Stop()
	engine.Start()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs
}

func TestServer2(t *testing.T) {
	//http://127.0.0.1:8222/connz?subs=true
	//nats pub identify.server.1.2.3 "hello world"
	engine := actor.NewEngine(0, 1, 1001, "")
	sServerAddress := concepts.GenServerAddress(0, 1, 1001)
	rpcServer := rpc.NewRPCServer(engine, "nats://127.0.0.1:4222", sServerAddress)
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
