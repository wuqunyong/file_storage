package concepts

import (
	"context"

	"github.com/wuqunyong/file_storage/pkg/encoders"
	"github.com/wuqunyong/file_storage/pkg/tick"
)

type IActor interface {
	ActorId() *ActorId
	Register(name string, fun interface{}) error
	Start()
	GetTimerQueue() *tick.TimerQueue
	Request(target *ActorId, method string, args any, opts ...context.Context) IMsgReq
	PostTask(funObj func()) error
	Stop()
	GetShutdownCh() <-chan struct{}
	Send(request IMsgReq) error
	IsRoot() bool
	Codec() encoders.IEncoder
	SetCodec(codec encoders.IEncoder)

	SpawnChild(actor IChildActor, id string) (*ActorId, error)

	OnInit() error
	OnShutdown()
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
