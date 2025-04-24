package actor

import (
	"context"
	"errors"
	"fmt"
	"runtime/debug"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"

	"github.com/wuqunyong/file_storage/pkg/concepts"
	"github.com/wuqunyong/file_storage/pkg/encoders"
	"github.com/wuqunyong/file_storage/pkg/errs"
	"github.com/wuqunyong/file_storage/pkg/logger"
	"github.com/wuqunyong/file_storage/pkg/msg"
	"github.com/wuqunyong/file_storage/pkg/tick"
)

type Actor struct {
	actorId      *concepts.ActorId
	context      *Context
	msgs         *Inbox
	tickDuration time.Duration
	timerQueue   *tick.TimerQueue
	codec        encoders.IEncoder

	shutdownCtx    context.Context
	shutdownCancel context.CancelFunc
	wg             sync.WaitGroup
	closed         atomic.Bool
	handler        concepts.IActorHandler
}

func NewActor(id string, e concepts.IEngine) *Actor {
	actorId := concepts.NewActorId(e.GetAddress(), id)
	a := &Actor{
		actorId:      actorId,
		msgs:         NewInbox(),
		tickDuration: time.Duration(10) * time.Millisecond,
		timerQueue:   tick.NewTimerQueue(),
		codec:        encoders.NewProtobufEncoder(),
	}
	a.closed.Store(false)

	shutdownCtx, shutdownCancel := context.WithCancel(context.Background())
	ctx := newContext(shutdownCtx, actorId, e)
	a.context = ctx
	a.context.parentCtx = nil
	a.shutdownCtx = shutdownCtx
	a.shutdownCancel = shutdownCancel
	return a
}

func NewChildActor(id string, e concepts.IEngine, parent *Context) *Actor {
	actorId := concepts.NewActorId(e.GetAddress(), id)
	a := &Actor{
		actorId:      actorId,
		msgs:         NewInbox(),
		tickDuration: time.Duration(10) * time.Millisecond,
		timerQueue:   tick.NewTimerQueue(),
		codec:        encoders.NewProtobufEncoder(),
	}
	a.closed.Store(false)

	shutdownCtx, shutdownCancel := context.WithCancel(parent.GetCtx())
	ctx := newContext(shutdownCtx, a.ActorId(), e)
	a.context = ctx
	a.context.parentCtx = parent
	a.shutdownCtx = shutdownCtx
	a.shutdownCancel = shutdownCancel
	return a
}

func (a *Actor) ActorId() *concepts.ActorId {
	return a.actorId
}

func (a *Actor) GetEngine() concepts.IEngine {
	return a.context.GetEngine()
}

func (a *Actor) Init() error {
	logger.Log(logger.DebugLevel, "Actor Init", "actorId", a.actorId.String(), "address", a.GetObjAddress())

	if a.handler != nil {
		return a.handler.OnInit()
	}
	return nil
}

func (a *Actor) Start() {
	a.wg.Add(1)
	go a.handleMsg()
}

func (a *Actor) Codec() encoders.IEncoder {
	return a.codec
}

func (a *Actor) SetCodec(codec encoders.IEncoder) {
	a.codec = codec
}

func (a *Actor) GetTimerQueue() *tick.TimerQueue {
	return a.timerQueue
}

func SendRequest[T any](actor concepts.IActor, target *concepts.ActorId, opcode uint32, args any, opts ...context.Context) (*T, errs.CodeError) {
	request := actor.Request(target, opcode, args, opts...)
	resp, err := msg.GetResult[T](request)
	return resp, err
}

func SendNotify(actor concepts.IActor, target *concepts.ActorId, opcode uint32, args any, opts ...context.Context) error {
	request := actor.Request(target, opcode, args, opts...)
	return request.Error()
}

func (a *Actor) Request(target *concepts.ActorId, opcode uint32, args any, opts ...context.Context) concepts.IMsgReq {
	var ctx context.Context
	if len(opts) > 0 {
		ctx = opts[0]
	}
	request := msg.NewMsgReq(target, opcode, args, ctx, a.Codec())
	if request.Err != nil {
		return request
	}

	if a.actorId.Equals(target) {
		request.Err = errors.New("forwarding message to self")
		return request
	}

	request.Sender = a.ActorId()
	err := a.context.engine.Request(request)
	if err != nil {
		request.Err = err
		return request
	}

	return request
}

func (a *Actor) Shutdown() {
	logger.Log(logger.DebugLevel, "Actor OnShutdown", "actorId", a.actorId.String(), "address", a.GetObjAddress())

	if a.handler != nil {
		a.handler.OnShutdown()
	}
}

func (a *Actor) Stop() {
	if a.closed.Load() {
		return
	}
	a.closed.Store(true)

	a.CallShutdown()
	a.StopChildren()

	a.waitForChildrenClosed()
	a.Shutdown()

	if a.context.GetParentCtx() != nil {
		a.context.GetParentCtx().children.Delete(a.actorId.ID)
	}

	a.context.engine.RemoveActor(a.ActorId())
}

func (a *Actor) waitForChildrenClosed() {
	for {
		if len(a.context.Children()) == 0 {
			break
		}
		time.Sleep(1 * time.Second)
	}
}

func (a *Actor) Register(opcode uint32, fun interface{}) error {
	return a.msgs.Register(opcode, fun)
}

func (a *Actor) Send(request concepts.IMsgReq) error {
	return a.msgs.Send(request)
}

func (a *Actor) PostTask(funObj func()) error {
	return a.msgs.Send(funObj)
}

func (a *Actor) IsRoot() bool {
	return a.context.Parent() == nil
}

func (a *Actor) GetParentActor() concepts.IActor {
	parent := a.context.Parent()
	if parent != nil {
		return a.context.engine.GetRegistry().GetByID(parent.ID)
	}
	return nil
}

func (a *Actor) handleMsg() {
	ticker := time.NewTicker(a.tickDuration)
	defer func() {
		ticker.Stop()
		if r := recover(); r != nil {
			errInfo := fmt.Sprint("task have panic err:", r, string(debug.Stack()))
			logger.Log(logger.ErrorLevel, "Actor handleMsg", "err", errInfo)
		}
		a.Stop()
	}()

	a.wg.Done()

	for {
		bDone := false
		select {
		case <-a.msgs.pendingCh:
			a.msgs.Run()
		case <-ticker.C:
			// fmt.Println("tick")
		case <-a.shutdownCtx.Done():
			bDone = true
		}

		if bDone {
			break
		}

		var lRestore []*tick.Timer

		iTime := time.Now().UnixNano()
		iLen := a.timerQueue.Len()
		for i := 0; i < iLen; i++ {
			timer := a.timerQueue.Peek()
			if iTime < timer.GetExpireTime() {
				break
			}
			timer = a.timerQueue.Pop()
			timer.Run()

			if !timer.IsOneshot() {
				timer.Restore()
				lRestore = append(lRestore, timer)
			}
		}

		for _, timer := range lRestore {
			a.timerQueue.Restore(timer)
		}
	}
}

func (a *Actor) CallShutdown() {
	a.shutdownCancel()
}

func (a *Actor) StopChildren() {
	children := a.context.Children()
	for _, child := range children {
		actorObj := a.GetEngine().GetRegistry().GetByID(child.GetId())
		if actorObj != nil {
			actorObj.CallShutdown()
			actorObj.StopChildren()
		}
	}

	a.wg.Wait()
}

func (a *Actor) SpawnChild(actor concepts.IChildActor, id string) (*concepts.ActorId, error) {
	return a.context.SpawnChild(actor, id)
}

func (a *Actor) FindChild(id string) *concepts.ActorId {
	return a.context.Child(id)
}

func (a *Actor) GetObjAddress() uintptr {
	return uintptr(unsafe.Pointer(a))
}

func (a *Actor) SetActorHandler(handler concepts.IActorHandler) {
	a.handler = handler
}

func (a *Actor) OnInit() error {
	return nil
}

func (a *Actor) OnShutdown() {
}
