package concepts

import "context"

type IActor interface {
	ActorId() *ActorId
	Register(name string, fun interface{}) error
	Init() error
	Start()
	Request(target *ActorId, method string, args any, opts ...context.Context) IMsgReq
	PostTask(funObj func()) error
	Stop()

	Send(request IMsgReq) error
}
