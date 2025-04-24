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
	Register(opcode uint32, fun interface{}) error

	Init() error
	Start()
	Stop()
	Shutdown()

	GetTimerQueue() *tick.TimerQueue
	Request(target *ActorId, opcode uint32, args any, opts ...context.Context) IMsgReq
	PostTask(funObj func()) error
	Send(request IMsgReq) error
	IsRoot() bool
	Codec() encoders.IEncoder
	SetCodec(codec encoders.IEncoder)

	SpawnChild(actor IChildActor, id string) (*ActorId, error)
	FindChild(id string) *ActorId

	CallShutdown()
	StopChildren()

	SetActorHandler(handler IActorHandler)

	GetObjAddress() uintptr
}

type IActorLoader interface {
	SetEmbeddingActor(actor IActor)
}

type IChildActor interface {
	IActor
	IActorLoader
}
