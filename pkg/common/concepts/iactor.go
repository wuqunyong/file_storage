package concepts

import (
	"context"

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

	SetEmbeddingActor(actor IActor)
	SpawnChild(actor IActor, id string) (*ActorId, error)

	OnInit() error
	OnShutdown()
}
