package rpc

import (
	"errors"
	"fmt"
	"runtime/debug"
	"sync/atomic"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/wuqunyong/file_storage/pkg/cluster"
	"github.com/wuqunyong/file_storage/pkg/common/concepts"
	"github.com/wuqunyong/file_storage/pkg/msg"
)

type RPCServer struct {
	connString             string
	connectionTimeout      time.Duration
	maxReconnectionRetries int
	conn                   *nats.Conn
	dieChan                chan bool
	topic                  *cluster.NatsSubject
	closed                 atomic.Bool
	engine                 concepts.IEngine
}

type RPCServerOpt func(*RPCServer)

func NewRPCServer(engine concepts.IEngine, connString, subjectName string, opts ...RPCServerOpt) *RPCServer {
	rpcServer := &RPCServer{
		engine:                 engine,
		connString:             connString,
		connectionTimeout:      time.Duration(6) * time.Second,
		maxReconnectionRetries: 3,
		dieChan:                make(chan bool),
		topic:                  cluster.NewNatsSubject(subjectName, 1024),
	}
	rpcServer.closed.Store(false)

	for _, opt := range opts {
		opt(rpcServer)
	}

	return rpcServer
}

func (rpc *RPCServer) Init() error {
	conn, err := cluster.SetupNatsConn(
		rpc.connString,
		rpc.dieChan,
		nats.MaxReconnects(rpc.maxReconnectionRetries),
		nats.Timeout(rpc.connectionTimeout),
	)
	if err != nil {
		return err
	}
	rpc.conn = conn

	rpc.topic.Subscription, err = rpc.conn.ChanSubscribe(rpc.topic.Subject, rpc.topic.Ch)
	if err != nil {
		fmt.Printf("[remoteProcess] Subscribe fail. [subject = %s, err = %s]", rpc.topic.Subject, err)
		return err
	}

	return nil
}

func (rpc *RPCServer) Run() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("rpc have panic err:", r, string(debug.Stack()))
		}
		rpc.Stop()
	}()

	process := func(natsMsg *nats.Msg) {
		fmt.Printf("msg:%+v\n", natsMsg)
		fmt.Printf("data:%s\n", string(natsMsg.Data))

		if len(natsMsg.Data) == 0 {
			return
		}

		request, err := msg.RequestUnmarshal(natsMsg.Data)
		if err != nil {
			fmt.Printf("err:%+v\n", err)
			return
		}
		rpc.HandleRequest(request)
		//reply := append(natsMsg.Data, "|response"...)
		//rpc.conn.Publish(natsMsg.Reply, reply)
	}

	for {
		select {
		case msg, ok := <-rpc.topic.Ch:
			if !ok {
				return
			}
			process(msg)
		case <-rpc.dieChan:
			return
		}
	}
}

func (rpc *RPCServer) HandleRequest(req concepts.IMsgReq) error {
	request, ok := req.(*msg.MsgReq)
	if !ok {
		return errors.New("invalid")
	}
	request.RPCServer = rpc
	rpc.engine.Request(request)
	return nil
}

func (rpc *RPCServer) SendResponse(subj string, response concepts.IMsgResp) error {
	if rpc.closed.Load() {
		return errors.New("rpc client has closed")
	}

	data, err := response.Marshal()
	if err != nil {
		return err
	}
	return rpc.conn.Publish(subj, data)
}

func (rpc *RPCServer) Stop() {
	if rpc.closed.Load() {
		return
	}
	rpc.closed.Store(true)

	rpc.topic.Stop()
	rpc.conn.Close()
}
