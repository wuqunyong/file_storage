package actor

import (
	"errors"
	"fmt"
	"sort"
	"sync"

	"github.com/wuqunyong/file_storage/pkg/common/concepts"
	"github.com/wuqunyong/file_storage/pkg/rpc"
)

type Engine struct {
	registry   *Registry
	address    string
	rpcFlag    bool
	rpcClient  concepts.IRPCClient
	rpcServer  concepts.IRPCServer
	mu         sync.Mutex
	components map[string]concepts.IComponent
	id2Name    map[string]string
}

type IComponentSlice []concepts.IComponent

func (c IComponentSlice) Len() int           { return len(c) }
func (c IComponentSlice) Less(i, j int) bool { return c[i].Priority() < c[j].Priority() }
func (c IComponentSlice) Swap(i, j int)      { c[i], c[j] = c[j], c[i] }

func NewEngine(kind, address string, rpcFlag bool, connString string) *Engine {
	sServerAddress := GenServerAddress(kind, address)
	e := &Engine{
		address:    sServerAddress,
		components: make(map[string]concepts.IComponent),
		id2Name:    make(map[string]string),
	}
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

func (e *Engine) AddComponent(component concepts.IComponent) error {
	if e.HasComponent(component.Name()) {
		sError := fmt.Sprintf("duplicate component name:%s", component.Name())
		return errors.New(sError)
	}

	if e.HasActor(component.ActorId()) {
		sError := fmt.Sprintf("duplicate actor id:%s", component.ActorId().String())
		return errors.New(sError)
	}

	name := component.Name()
	id := component.ActorId().GetId()
	e.mu.Lock()
	defer e.mu.Unlock()
	_, err := e.SpawnActor(component)
	if err != nil {
		return err
	}
	e.components[name] = component
	e.id2Name[id] = name

	return nil
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
		sError := fmt.Sprintf("not exist:%v\n", request.GetTarget().Address)
		return errors.New(sError)
	}
	return actorObj.Send(request)
}

func (e *Engine) Init() error {
	var err error
	if e.rpcFlag {
		err := e.rpcClient.Init()
		if err != nil {
			return err
		}
		err = e.rpcServer.Init()
		if err != nil {
			return err
		}
	}

	var componentsObj IComponentSlice
	for _, value := range e.components {
		componentsObj = append(componentsObj, value)
	}
	sort.Sort(componentsObj)

	for _, obj := range componentsObj {
		err := obj.OnInit()
		if err != nil {
			return err
		}
	}

	return err
}

func (e *Engine) Start() {
	if e.rpcFlag {
		go e.rpcClient.Run()
		go e.rpcServer.Run()
	}
}

func (e *Engine) Stop() {
	var componentsObj IComponentSlice
	for _, value := range e.components {
		componentsObj = append(componentsObj, value)
	}
	sort.Sort(sort.Reverse(componentsObj))
	for _, obj := range componentsObj {
		obj.OnCleanup()
	}

	if e.rpcFlag {
		e.rpcClient.Stop()
		e.rpcServer.Stop()
	}
}

func (e *Engine) isLocalMessage(actor *concepts.ActorId) bool {
	return e.address == actor.Address
}
