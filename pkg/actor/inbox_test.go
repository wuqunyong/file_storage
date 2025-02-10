package actor

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/wuqunyong/file_storage/pkg/encoders"
	"github.com/wuqunyong/file_storage/pkg/errs"
	"github.com/wuqunyong/file_storage/pkg/msg"
	testdata "github.com/wuqunyong/file_storage/protobuf"
)

type Reply1 struct {
	Value int32
}

type Handler struct {
}

var inboxObj *Inbox = NewInbox()

func (h *Handler) Func1(ctx context.Context, arg *testdata.Person, reply *testdata.Person) errs.CodeError {
	reply.Age += arg.Age
	reply.Name = "Func1"
	fmt.Printf("inside value:%v\n", reply)

	request := &testdata.Person{Name: "小明", Age: 18}
	decoder := encoders.NewProtobufEncoder()
	data, _ := decoder.Encode(request)
	msgReq := &msg.MsgReq{
		SeqId:    2,
		FuncName: 2,
		ArgsData: data,
		Done:     make(chan *msg.MsgResp),
		Ctx:      context.Background(),
	}
	inboxObj.Send(msgReq)
	response, _ := msgReq.Result()
	fmt.Printf("Func2 %v %T", response, response)

	return errs.NewCodeError(errors.New("invalid"))
}

func (h *Handler) Func2(ctx context.Context, arg *testdata.Person, reply *testdata.Person) errs.CodeError {
	reply.Age += arg.Age
	reply.Name = "Func2"
	fmt.Printf("inside value:%v\n", reply)
	return errs.NewCodeError(errors.New("invalid"))
}

func Test1(t *testing.T) {
	handler := Handler{}

	inboxObj.Register(1, handler.Func1)
	inboxObj.Register(2, handler.Func2)
	go func() {
		for {
			select {
			case <-inboxObj.pendingCh:
				inboxObj.Run()
			}
		}
	}()

	reply := &testdata.Person{Name: "小明", Age: 18}
	decoder := encoders.NewProtobufEncoder()
	data, err := decoder.Encode(reply)
	if err != nil {
		t.Fatal("Couldn't serialize object", err)
	}

	msgReq := &msg.MsgReq{
		SeqId:    1,
		FuncName: 1,
		ArgsData: data,
		Done:     make(chan *msg.MsgResp),
		Ctx:      context.Background(),
	}
	err = inboxObj.Send(msgReq)
	if err != nil {
		t.Fatal("err", err)
	}

	response, _ := msgReq.Result()
	fmt.Printf("%v %T", response, response)

	time.Sleep(time.Duration(60) * time.Second)
	time.Sleep(time.Duration(60) * time.Second)
}
