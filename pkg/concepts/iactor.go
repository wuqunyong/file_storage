package concepts

import (
	"context"

	"github.com/wuqunyong/file_storage/pkg/encoders"
	"github.com/wuqunyong/file_storage/pkg/tick"
)

type IActorHandler interface {
	OnInit() error
	OnShutdown()
}

type IActor interface {
	IActorHandler
	ActorId() *ActorId
	Register(name string, fun interface{}) error

	Start()
	Stop()
	GetShutdownCh() <-chan struct{}

	GetTimerQueue() *tick.TimerQueue
	Request(target *ActorId, method string, args any, opts ...context.Context) IMsgReq
	PostTask(funObj func()) error
	Send(request IMsgReq) error
	IsRoot() bool
	Codec() encoders.IEncoder
	SetCodec(codec encoders.IEncoder)

	SpawnChild(actor IChildActor, id string) (*ActorId, error)
	FindChild(id string) *ActorId

	SetActorHandler(handler IActorHandler)

	Init() error
	Shutdown()

	GetObjAddress() uintptr
}

type IActorLoader interface {
	SetEmbeddingActor(actor IActor)
}

type IChildActor interface {
	IActor
	IActorLoader
}

type ChildActor struct {
	IActor
}

func (child *ChildActor) SetEmbeddingActor(actor IActor) {
	child.IActor = actor
}
