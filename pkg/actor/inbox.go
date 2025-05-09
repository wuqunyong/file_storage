package actor

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"runtime/debug"
	"sync"

	"github.com/wuqunyong/file_storage/pkg/encoders"
	"github.com/wuqunyong/file_storage/pkg/errs"
	"github.com/wuqunyong/file_storage/pkg/funcutils"
	"github.com/wuqunyong/file_storage/pkg/logger"
	"github.com/wuqunyong/file_storage/pkg/msg"
	"github.com/wuqunyong/file_storage/pkg/queue"
)

type Inbox struct {
	lock   sync.Mutex
	method map[uint32]*funcutils.MethodType // registered methods

	pending   *queue.MsgQueue
	pendingCh chan struct{}
	ctx       context.Context
}

func NewInbox() *Inbox {
	inbox := &Inbox{
		method:    make(map[uint32]*funcutils.MethodType),
		pending:   queue.NewTimerQueue(),
		pendingCh: make(chan struct{}, 1),
		ctx:       context.Background(),
	}
	return inbox
}

func (inbox *Inbox) Register(opcode uint32, fun interface{}) error {
	inbox.lock.Lock()
	defer inbox.lock.Unlock()

	_, ok := inbox.method[opcode]
	if ok {
		return errors.New(fmt.Sprintf("duplicate name:%d", opcode))
	}

	ptrMethod := funcutils.GetRPCReflectFunc(fun, true)
	if ptrMethod == nil {
		return errors.New(fmt.Sprintf("GetRPCReflectFunc err:%d", opcode))
	}

	inbox.method[opcode] = ptrMethod
	return nil
}

func (inbox *Inbox) SetContext(ctx context.Context) error {
	if ctx == nil {
		return errors.New("ctx is nil")
	}

	inbox.ctx = ctx
	return nil
}

func (inbox *Inbox) Send(args any) error {
	switch message := args.(type) {
	case *msg.MsgReq:
		inbox.SendMsgReq(args)
	case func():
		inbox.SendMsgReq(args)
	default:
		sError := fmt.Sprintf("unexpected type:%T", message)
		return errors.New(sError)
	}

	return nil
}

func (inbox *Inbox) SendMsgReq(args any) {
	inbox.pending.Push(args)
	select {
	case inbox.pendingCh <- struct{}{}:
		// ok
	default:
		// nothing
	}
}

func (inbox *Inbox) Run() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Inbox have panic err:", r, string(debug.Stack()))
		}
		inbox.Stop()
	}()

	count := inbox.pending.Len()
	for i := 0; i < count; i++ {
		item := inbox.pending.Pop()
		if item == nil {
			break
		}

		switch message := item.Args.(type) {
		case *msg.MsgReq:
			inbox.handleMsgReq(message)
		case func():
			inbox.handleFuncObj(message)
		default:
			continue
		}
	}
}

func (inbox *Inbox) handleMsgReq(message *msg.MsgReq) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("handleMsgReq have panic err:", r, string(debug.Stack()))
		}
	}()
	response := inbox.callFunc(message)
	if response != nil {
		go message.Send(response)
	}
}

func (inbox *Inbox) handleFuncObj(funcObj func()) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("handleFuncObj have panic err:", r, string(debug.Stack()))
		}
	}()
	funcObj()
}

func (inbox *Inbox) callFunc(message *msg.MsgReq) *msg.MsgResp {
	inbox.lock.Lock()
	defer inbox.lock.Unlock()

	ptrMethod, ok := inbox.method[message.FuncName]
	if !ok {
		sError := fmt.Sprintf("unregister name:%d", message.FuncName)
		reply := msg.NewMsgResp(message.SeqId, 1, sError, message.Codec)
		return reply
	}

	var (
		code    errs.CodeError
		err     error
		decoder = encoders.NewProtobufEncoder()
	)

	if message.Remote {
		args := reflect.New(ptrMethod.ArgType[1].Elem()).Interface()
		err := decoder.Decode(message.ArgsData, args)
		if err != nil {
			sError := fmt.Sprintf("Decode err:%v", err)
			reply := msg.NewMsgResp(message.SeqId, 1, sError, message.Codec)
			return reply
		}

		switch ptrMethod.NumIn {
		case 3:
			var response = reflect.New(ptrMethod.ReplyType.Elem()).Interface()
			code, err = funcutils.CallPRCReflectRequestFunc(ptrMethod, inbox.ctx, args, response)
			if err != nil {
				sError := fmt.Sprintf("err:%s" + err.Error())
				reply := msg.NewMsgResp(message.SeqId, 1, sError, message.Codec)
				return reply
			}

			if code != nil {
				reply := msg.NewMsgResp(message.SeqId, uint32(code.Code()), code.Msg(), message.Codec)
				return reply
			}

			replyData, err := decoder.Encode(response)
			if err != nil {
				sError := fmt.Sprintf("Encode err:%v", err)
				reply := msg.NewMsgResp(message.SeqId, 1, sError, message.Codec)
				return reply
			}

			reply := msg.NewMsgResp(message.SeqId, 0, "", message.Codec)
			reply.Remote = message.Remote
			reply.ReplyData = replyData
			return reply
		case 2:
			err = funcutils.CallPRCReflectNotifyFunc(ptrMethod, inbox.ctx, args)
			if err != nil {
				sError := fmt.Sprintf("err:%s", err.Error())
				logger.Log(logger.InfoLevel, "callFunc", "Error", sError)
			}

			return nil
		default:
			sError := fmt.Sprintf("err:invalid ptrMethod.NumIn:%d", ptrMethod.NumIn)
			logger.Log(logger.InfoLevel, "callFunc", "Error", sError)
			return nil
		}
	}

	switch ptrMethod.NumIn {
	case 3:
		var response = reflect.New(ptrMethod.ReplyType.Elem()).Interface()
		code, err = funcutils.CallPRCReflectRequestFunc(ptrMethod, inbox.ctx, message.Args, response)
		if err != nil {
			sError := fmt.Sprintf("err:%s" + err.Error())
			reply := msg.NewMsgResp(message.SeqId, 1, sError, message.Codec)
			return reply
		}
		if code != nil {
			reply := msg.NewMsgResp(message.SeqId, uint32(code.Code()), code.Msg(), message.Codec)
			return reply
		}

		reply := msg.NewMsgResp(message.SeqId, 0, "", message.Codec)
		reply.Remote = message.Remote
		reply.Reply = response
		return reply
	}

	return nil
}

func (inbox *Inbox) Stop() {

}
