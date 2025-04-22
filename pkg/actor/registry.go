package actor

import (
	"errors"
	"fmt"
	"sync"

	"github.com/wuqunyong/file_storage/pkg/concepts"
	"github.com/wuqunyong/file_storage/pkg/logger"
)

const LocalLookupAddr = "local"

type Registry struct {
	mu     sync.RWMutex
	lookup map[string]concepts.IActor
	root   map[string]bool
	engine concepts.IEngine
}

func newRegistry(e concepts.IEngine) *Registry {
	return &Registry{
		lookup: make(map[string]concepts.IActor, 1024),
		root:   make(map[string]bool),
		engine: e,
	}
}

// // GetPID returns the process id associated for the given kind and its id.
// // GetPID returns nil if the process was not found.
// func (r *Registry) GetActorId(kind, id string) *concepts.ActorId {
// 	actor := r.GetByID(kind + "." + id)
// 	if actor != nil {
// 		return actor.ActorId()
// 	}
// 	return nil
// }

// Remove removes the given PID from the registry.
func (r *Registry) Remove(actorId *concepts.ActorId) {
	logger.Log(logger.InfoLevel, "Actor Remove", "actorId", actorId.String())

	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.lookup, actorId.ID)
	delete(r.root, actorId.ID)
}

// get returns the processer for the given PID, if it exists.
// If it doesn't exist, nil is returned so the caller must check for that
// and direct the message to the deadletter processer instead.
func (r *Registry) get(pid *concepts.ActorId) concepts.IActor {
	if pid == nil {
		return nil
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	if actor, ok := r.lookup[pid.ID]; ok {
		return actor
	}
	return nil
}

func (r *Registry) GetByID(id string) concepts.IActor {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.lookup[id]
}

func (r *Registry) GetRootID() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	ids := make([]string, 0, len(r.root))
	for key := range r.root {
		ids = append(ids, key)
	}
	return ids
}

func (r *Registry) add(actor concepts.IActor) error {
	logger.Log(logger.InfoLevel, "Actor add", "actorId", actor.ActorId().String())

	r.mu.Lock()
	id := actor.ActorId().ID
	if _, ok := r.lookup[id]; ok {
		r.mu.Unlock()
		sError := fmt.Sprintf("duplicate actor id:%s", id)
		return errors.New(sError)
	}
	r.lookup[id] = actor
	r.mu.Unlock()
	err := actor.Init()
	if err != nil {
		return err
	}
	actor.Start()
	if actor.IsRoot() {
		r.mu.Lock()
		r.root[id] = true
		r.mu.Unlock()
	}
	return nil
}
