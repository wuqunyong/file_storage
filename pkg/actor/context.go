package actor

import (
	"context"

	"github.com/wuqunyong/file_storage/pkg/common/concepts"
)

type Context struct {
	engine  concepts.IEngine
	context context.Context
}

func newContext(ctx context.Context, e concepts.IEngine) *Context {
	return &Context{
		context: ctx,
		engine:  e,
	}
}
