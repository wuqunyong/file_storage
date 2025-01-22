package actor

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"runtime/debug"
	"sync/atomic"
	"time"
	"unsafe"

	"github.com/wuqunyong/file_storage/pkg/concepts"
	"github.com/wuqunyong/file_storage/pkg/encoders"
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

	cancelch   chan struct{}
	shutdownCh chan struct{}
	closed     atomic.Bool
	handler    concepts.IActorHandler
}

func NewActor(id string, e concepts.IEngine) *Actor {
	actorId := concepts.NewActorId(e.GetAddress(), id)
	ctx := newContext(context.Background(), actorId, e)
	a := &Actor{
		actorId:      actorId,
		context:      ctx,
		msgs:         NewInbox(),
		tickDuration: time.Duration(10) * time.Millisecond,
		timerQueue:   tick.NewTimerQueue(),
		codec:        encoders.NewProtobufEncoder(),
		cancelch:     make(chan struct{}),
		shutdownCh:   make(chan struct{}),
	}
	a.closed.Store(false)
	return a
}

func (a *Actor) ActorId() *concepts.ActorId {
	return a.actorId
}

func (a *Actor) GetEngine() concepts.IEngine {
	return a.context.GetEngine()
}

func (a *Actor) Init() error {
	if a.handler != nil {
		return a.handler.OnInit()
	}
	return nil
}

func (a *Actor) Start() {
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

func (a *Actor) Request(target *concepts.ActorId, method string, args any, opts ...context.Context) concepts.IMsgReq {
	var ctx context.Context
	if len(opts) > 0 {
		ctx = opts[0]
	}
	request := msg.NewMsgReq(target, method, args, ctx, a.Codec())
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
	slog.Info("Actor OnShutdown", "actorId", a.actorId.String(), "address", a.GetObjAddress())
	if a.handler != nil {
		a.handler.OnShutdown()
	}
}

func (a *Actor) Stop() {
	if a.closed.Load() {
		return
	}
	a.closed.Store(true)

	close(a.shutdownCh)
	a.waitForChildrenClosed()
	a.Shutdown()

	if a.context.GetParentCtx() != nil {
		a.context.GetParentCtx().children.Delete(a.actorId.ID)
		slog.Info("Actor Delete", "actorId", a.actorId.ID)
	}

	close(a.cancelch)
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

func (a *Actor) Register(name string, fun interface{}) error {
	return a.msgs.Register(name, fun)
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

func (a *Actor) GetShutdownCh() <-chan struct{} {
	return a.shutdownCh
}

func (a *Actor) GetParentShutdownCh() <-chan struct{} {
	parent := a.context.Parent()
	if parent != nil {
		for {
			parent := a.GetParentActor()
			if parent != nil {
				return parent.GetShutdownCh()
			}

			time.Sleep(10 * time.Millisecond)
		}
	}

	tmpCh := make(chan struct{})
	return tmpCh
}

func (a *Actor) handleMsg() {
	ticker := time.NewTicker(a.tickDuration)
	defer func() {
		ticker.Stop()
		if r := recover(); r != nil {
			fmt.Println("task have panic err:", r, string(debug.Stack()))
		}
		a.Stop()
	}()

	parentShutdownCh := a.GetParentShutdownCh()
	for {
		bDone := false
		select {
		case <-a.msgs.pendingCh:
			a.msgs.Run()
		case <-ticker.C:
			// fmt.Println("tick")
		case <-a.cancelch:
			bDone = true
		case <-parentShutdownCh:
			bDone = true
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

		if bDone {
			break
		}
	}
}

func (a *Actor) SpawnChild(actor concepts.IChildActor, id string) (*concepts.ActorId, error) {
	return a.context.SpawnChild(actor, id)
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
