package actor

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"testing"
	"time"

	"github.com/wuqunyong/file_storage/pkg/errs"
	"github.com/wuqunyong/file_storage/pkg/logger"
	"github.com/wuqunyong/file_storage/pkg/msg"
	"github.com/wuqunyong/file_storage/proto/common_msg"
	"google.golang.org/protobuf/proto"
)

type ActorObjA struct {
	*Actor
}

func (actor *ActorObjA) OnInit() error {
	actor.Register(1, actor.Func1)
	actor.Register(2, actor.Func2)
	return nil
}

func (actor *ActorObjA) OnShutdown() {
}

func (actor *ActorObjA) Func1(ctx context.Context, request *common_msg.EchoRequest, reply *common_msg.EchoResponse) errs.CodeError {
	reply.Value1 = request.Value1 + 1
	reply.Value2 = request.Value2 + " response"

	logger.Log(logger.InfoLevel, "Func1", "request", request, "reply", reply)

	return nil
}

func (actor *ActorObjA) Func2(ctx context.Context, request *common_msg.EchoRequest, reply *common_msg.EchoResponse) errs.CodeError {
	reply.Value1 = request.Value1 + 1
	reply.Value2 = request.Value2 + " response"

	logger.Log(logger.InfoLevel, "Func2", "request", request, "reply", reply)

	return errs.NewCodeError(errors.New("invalid"), 123)
}

func Must[T proto.Message](arg []byte, object T) T {
	err := proto.Unmarshal(arg, object)
	if err != nil {
		panic(err)
	}
	return object
}

type ChildObjA struct {
	ChildActor
	id int
}

func (actor *ChildObjA) OnInit() error {
	logger.Log(logger.InfoLevel, "ChildObjA OnInit", "id", actor.id)
	return nil
}

func (actor *ChildObjA) OnShutdown() {
	logger.Log(logger.InfoLevel, "ChildObjA OnShutdown", "id", actor.id)
}

func Test(t *testing.T) {

	engine := NewEngine(0, 1, 1001)
	actorObj1 := &ActorObjA{
		Actor: NewActor("1", engine),
	}
	actorObj2 := &ActorObjA{
		Actor: NewActor("2", engine),
	}

	engine.MustInit()
	engine.SpawnActor(actorObj1)
	engine.SpawnActor(actorObj2)

	for i := 0; i < 3; i++ {
		childObj := &ChildObjA{id: i}
		actorObj1.SpawnChild(childObj, fmt.Sprintf("child.%d", i))
	}

	for i := 4; i < 7; i++ {
		childObj := &ChildObjA{id: i}
		actorObj2.SpawnChild(childObj, fmt.Sprintf("child.%d", i))
	}

	engine.Start()

	for i := 0; i < 3; i++ {
		echo := &common_msg.EchoRequest{Value1: 123456, Value2: "小明"}
		request := actorObj1.Request(actorObj2.ActorId(), 1, echo)
		obj, err := msg.GetResult[common_msg.EchoResponse](request)
		if err != nil {
			t.Fatal("DecodeResponse", err)
		}

		logger.Log(logger.InfoLevel, "actorObj1 EchoRequest", "obj", obj, "err", err)
	}

	for i := 0; i < 3; i++ {
		person := &common_msg.EchoRequest{Value1: 123456, Value2: "小明"}
		request := actorObj2.Request(actorObj1.ActorId(), 2, person)
		obj, err := msg.GetResult[common_msg.EchoResponse](request)
		if err != nil {
			sError := fmt.Sprintf("DecodeResponse: %v\n", err)
			logger.Log(logger.ErrorLevel, "actorObj2 EchoRequest", "err", sError)
			break
		}
		logger.Log(logger.InfoLevel, "actorObj2 EchoRequest", "obj", obj, "err", err)
	}

	time.Sleep(time.Duration(6) * time.Second)

	engine.Stop()

	time.Sleep(time.Duration(6) * time.Second)

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs
}
