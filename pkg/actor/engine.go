package actor

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"sort"
	"sync"
	"syscall"
	"time"

	"github.com/wuqunyong/file_storage/pkg/concepts"
	"github.com/wuqunyong/file_storage/pkg/rpc"
)

type ServerState int

const (
	STATE_UNINITIALIZED ServerState = iota
	STATE_INITIALIZED
	STATE_RUNNING
	STATE_SHUTTING_DOWN
	STATE_SHUTDOWN
)

type Engine struct {
	registry   *Registry
	address    string
	connString string
	rpcFlag    bool
	rpcClient  concepts.IRPCClient
	rpcServer  concepts.IRPCServer
	state      ServerState
	lastError  error
	mu         sync.Mutex
	components map[string]concepts.IComponent
}

type IComponentSlice []concepts.IComponent

func (c IComponentSlice) Len() int           { return len(c) }
func (c IComponentSlice) Less(i, j int) bool { return c[i].Priority() < c[j].Priority() }
func (c IComponentSlice) Swap(i, j int)      { c[i], c[j] = c[j], c[i] }

func NewEngine(kind, address string, connString string) *Engine {
	sServerAddress := GenServerAddress(kind, address)
	rpcFlag := false
	if connString != "" {
		rpcFlag = true
	}
	e := &Engine{
		address:    sServerAddress,
		connString: connString,
		components: make(map[string]concepts.IComponent),
	}
	e.setState(STATE_UNINITIALIZED)
	e.registry = newRegistry(e)
	e.rpcFlag = rpcFlag
	if e.rpcFlag {
		sClientAddress := GenClientAddress(kind, address)
		e.rpcClient = rpc.NewRPCClient(e, connString, sClientAddress)
		e.rpcServer = rpc.NewRPCServer(e, connString, sServerAddress)
	}
	return e
}

func GenServerAddress(kind, address string) string {
	sAddress := fmt.Sprintf("engine.%s.server.%s", kind, address)
	return sAddress
}

func GenClientAddress(kind, address string) string {
	sAddress := fmt.Sprintf("engine.%s.client.%s", kind, address)
	return sAddress
}

func (e *Engine) GetRegistry() concepts.IRegistry {
	return e.registry
}

func (e *Engine) MustAddComponent(component concepts.IComponent) {
	if e.HasComponent(component.Name()) {
		sError := fmt.Sprintf("duplicate component name:%s", component.Name())
		panic(sError)
	}

	name := component.Name()
	e.mu.Lock()
	defer e.mu.Unlock()
	e.components[name] = component
}
func (e *Engine) HasComponent(name string) bool {
	componentObj := e.GetComponent(name)
	return componentObj != nil
}

func (e *Engine) GetComponent(name string) concepts.IComponent {
	e.mu.Lock()
	defer e.mu.Unlock()
	componentObj, ok := e.components[name]
	if ok {
		return componentObj
	}
	return nil
}

func (e *Engine) GetAddress() string {
	return e.address
}

func (e *Engine) SpawnActor(actor concepts.IActor) (*concepts.ActorId, error) {
	slog.Info("Engine SpawnChild", "actorId", actor.ActorId().String(), "address", actor.GetObjAddress())
	actor.SetActorHandler(actor)
	err := e.registry.add(actor)
	if err != nil {
		return nil, err
	}

	return actor.ActorId(), nil
}

func (e *Engine) HasActor(id *concepts.ActorId) bool {
	actorObj := e.registry.get(id)
	return actorObj != nil
}

func (e *Engine) RemoveActor(id *concepts.ActorId) {
	e.registry.Remove(id)
}

func (e *Engine) Request(request concepts.IMsgReq) error {
	if !e.isLocalMessage(request.GetTarget()) {
		if !e.rpcFlag {
			return errors.New("rpcFlag is false")
		}
		err := request.SetRemote(true)
		if err != nil {
			return err
		}
		return e.rpcClient.SendRequest(request)
	}

	actorObj := e.registry.get(request.GetTarget())
	if actorObj == nil {
		sError := fmt.Sprintf("not exist:%v", request.GetTarget().Address)
		return errors.New(sError)
	}
	return actorObj.Send(request)
}

func (e *Engine) MustInit() {
	if e.rpcFlag {
		err := e.rpcClient.Init()
		if err != nil {
			e.lastError = err
			panic(err)
		}
		err = e.rpcServer.Init()
		if err != nil {
			e.lastError = err
			panic(err)
		}
	}

	components := e.GetComponentSlice(false)
	for _, obj := range components {
		obj.SetEngine(e)
		err := obj.OnInit()
		if err != nil {
			e.lastError = err
			panic(err)
		}
	}

	e.setState(STATE_INITIALIZED)
}

func (e *Engine) Start() {
	if e.rpcFlag {
		go e.rpcClient.Run()
		go e.rpcServer.Run()
	}

	components := e.GetComponentSlice(false)
	for _, obj := range components {
		obj.OnStart()
	}
	e.setState(STATE_RUNNING)
}

func (e *Engine) GetComponentSlice(reverse bool) IComponentSlice {
	var components IComponentSlice
	for _, value := range e.components {
		components = append(components, value)
	}
	if reverse {
		sort.Sort(sort.Reverse(components))
	} else {
		sort.Sort(components)
	}
	return components
}

func (e *Engine) waitForRootClosed() {
	for {
		rootIds := e.registry.GetRootID()
		if len(rootIds) == 0 {
			break
		}
		time.Sleep(1 * time.Second)
	}
}

func (e *Engine) Stop() {
	e.setState(STATE_SHUTTING_DOWN)
	rootIds := e.registry.GetRootID()
	for _, id := range rootIds {
		actor := e.registry.GetByID(id)
		if actor != nil {
			actor.Stop()
		}
	}
	e.waitForRootClosed()

	components := e.GetComponentSlice(true)
	for _, obj := range components {
		obj.OnCleanup()
	}

	if e.rpcFlag {
		e.rpcClient.Stop()
		e.rpcServer.Stop()
	}
	e.setState(STATE_SHUTDOWN)
}

func (e *Engine) setState(state ServerState) {
	e.state = state
	slog.Warn("Engine ChangeState", "state", e.state)
}

func (e *Engine) isLocalMessage(actor *concepts.ActorId) bool {
	return e.address == actor.Address
}

func (e *Engine) WaitForShutdown() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs

	progName := filepath.Base(os.Args[0])
	slog.Warn("Engine Exit", "reason", fmt.Sprintf("Warning %s receive process terminal SIGTERM exit 0", progName))
}
