package actor

import (
	"context"
	"errors"
	"fmt"
	"runtime/debug"
	"sync/atomic"
	"time"

	"github.com/wuqunyong/file_storage/pkg/common/concepts"
	"github.com/wuqunyong/file_storage/pkg/msg"
	"github.com/wuqunyong/file_storage/pkg/tick"
)

type Actor struct {
	actorId      *concepts.ActorId
	context      *Context
	msgs         *Inbox
	tickDuration time.Duration
	timerQueue   *tick.TimerQueue

	cancelch chan struct{}
	closed   atomic.Bool
}

func NewActor(id string, e concepts.IEngine) *Actor {
	actorId := concepts.NewActorId(e.GetAddress(), id)
	ctx := newContext(context.Background(), e)
	a := &Actor{
		actorId:      actorId,
		context:      ctx,
		msgs:         NewInbox(),
		tickDuration: time.Duration(10) * time.Millisecond,
		timerQueue:   tick.NewTimerQueue(),
		cancelch:     make(chan struct{}, 1),
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
	return nil
}

func (a *Actor) Start() {
	go a.handleMsg()
}

func (a *Actor) GetTimerQueue() *tick.TimerQueue {
	return a.timerQueue
}

func (a *Actor) Request(target *concepts.ActorId, method string, args any, opts ...context.Context) concepts.IMsgReq {
	var ctx context.Context
	if len(opts) > 0 {
		ctx = opts[0]
	}
	request := msg.NewMsgReq(target, method, args, ctx)
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

func (a *Actor) Stop() {
	if a.closed.Load() {
		return
	}
	a.closed.Store(true)
	close(a.cancelch)
	a.context.engine.RemoveActor(a.ActorId())
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

func (a *Actor) handleMsg() {
	ticker := time.NewTicker(a.tickDuration)
	defer func() {
		ticker.Stop()
		if r := recover(); r != nil {
			fmt.Println("task have panic err:", r, string(debug.Stack()))
		}
		a.Stop()
	}()

	for {
		bDone := false
		select {
		case <-a.msgs.pendingCh:
			a.msgs.Run()
		case <-ticker.C:
			// fmt.Println("tick")
		case <-a.cancelch:
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
