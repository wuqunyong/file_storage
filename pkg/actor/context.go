package actor

import (
	"context"
	"log/slog"

	"github.com/wuqunyong/file_storage/pkg/common/concepts"
	"github.com/wuqunyong/file_storage/pkg/safemap"
)

type Context struct {
	actorId   *concepts.ActorId
	engine    concepts.IEngine
	context   context.Context
	parentCtx *Context
	children  *safemap.SafeMap[string, *concepts.ActorId]
}

func newContext(actorId *concepts.ActorId, ctx context.Context, e concepts.IEngine) *Context {
	return &Context{
		actorId:  actorId,
		context:  ctx,
		engine:   e,
		children: safemap.New[string, *concepts.ActorId](),
	}
}

func (c *Context) GetEngine() concepts.IEngine {
	return c.engine
}

func (c *Context) GetParentCtx() *Context {
	return c.parentCtx
}

func (c *Context) GetActorId(id string) *concepts.ActorId {
	actor := c.engine.GetRegistry().GetByID(id)
	if actor != nil {
		return actor.ActorId()
	}
	return nil
}

// Parent returns the PID of the process that spawned the current process.
func (c *Context) Parent() *concepts.ActorId {
	if c.parentCtx != nil {
		return c.parentCtx.actorId
	}
	return nil
}

// Child will return the PID of the child (if any) by the given name/id.
// PID will be nil if it could not find it.
func (c *Context) Child(id string) *concepts.ActorId {
	pid, _ := c.children.Get(id)
	return pid
}

// Children returns all child PIDs for the current process.
func (c *Context) Children() []*concepts.ActorId {
	pids := make([]*concepts.ActorId, c.children.Len())
	i := 0
	c.children.ForEach(func(_ string, child *concepts.ActorId) {
		pids[i] = child
		i++
	})
	return pids
}

// PID returns the PID of the process that belongs to the context.
func (c *Context) ActorID() *concepts.ActorId {
	return c.actorId
}

func (c *Context) SpawnChild(actor concepts.IActor, id string) *concepts.ActorId {
	childId := c.actorId.GetId() + "." + id
	childActor := NewActor(childId, c.engine)
	childActor.context.parentCtx = c
	c.children.Set(childActor.actorId.ID, childActor.actorId)
	actor.SetEmbeddingActor(childActor)
	_, err := c.engine.SpawnActor(actor)
	if err != nil {
		slog.Error("SpawnChild", "err", err)
	}

	return childActor.actorId
}
