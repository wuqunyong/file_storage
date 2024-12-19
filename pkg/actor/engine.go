package actor

import (
	"errors"
	"fmt"

	"github.com/wuqunyong/file_storage/pkg/common/concepts"
	"github.com/wuqunyong/file_storage/pkg/rpc"
)

type Engine struct {
	Registry  *Registry
	address   string
	rpcFlag   bool
	rpcClient concepts.IRPCClient
	rpcServer concepts.IRPCServer
}

func NewEngine(kind, address string, rpcFlag bool, connString string) *Engine {
	sServerAddress := GenServerAddress(kind, address)
	e := &Engine{
		address: sServerAddress,
	}
	e.Registry = newRegistry(e)
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

func (e *Engine) SpawnActor(actor concepts.IActor) (*concepts.ActorId, error) {
	err := e.Registry.add(actor)
	if err != nil {
		return nil, err
	}
	return actor.ActorId(), nil
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

	actorObj := e.Registry.get(request.GetTarget())
	if actorObj == nil {
		sError := fmt.Sprintf("not exist:%v\n", request.GetTarget().Address)
		return errors.New(sError)
	}
	return actorObj.Send(request)
}

func (e *Engine) Init() error {
	if !e.rpcFlag {
		return nil
	}

	err := e.rpcClient.Init()
	if err != nil {
		return err
	}
	err = e.rpcServer.Init()
	return err
}

func (e *Engine) Start() {
	go e.rpcClient.Run()
	go e.rpcServer.Run()
}

func (e *Engine) Stop() {
	e.rpcClient.Stop()
	e.rpcServer.Stop()
}

func (e *Engine) isLocalMessage(actor *concepts.ActorId) bool {
	return e.address == actor.Address
}
