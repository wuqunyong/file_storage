package actor

import "context"

type Context struct {
	engine  *Engine
	context context.Context
}

func newContext(ctx context.Context, e *Engine) *Context {
	return &Context{
		context: ctx,
		engine:  e,
	}
}
