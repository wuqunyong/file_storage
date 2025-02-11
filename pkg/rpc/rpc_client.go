package rpc

import (
	"errors"
	"fmt"
	"log/slog"
	"runtime/debug"
	"sync"
	"sync/atomic"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/wuqunyong/file_storage/pkg/cluster"
	"github.com/wuqunyong/file_storage/pkg/concepts"
	"github.com/wuqunyong/file_storage/pkg/msg"
)

type RPCClient struct {
	id                     string
	connString             string
	connectionTimeout      time.Duration
	maxReconnectionRetries int
	conn                   *nats.Conn
	dieChan                chan bool
	topic                  *cluster.NatsSubject
	closed                 atomic.Bool
	seqId                  uint64
	pendingMu              sync.Mutex
	pending                map[uint64]concepts.IMsgReq
	engine                 concepts.IEngine
}

type RPCClientOpt func(*RPCClient)

func NewRPCClient(engine concepts.IEngine, connString, subjectName string, opts ...RPCClientOpt) *RPCClient {
	rpcClient := &RPCClient{
		id:                     fmt.Sprintf("RPCClient:%s", time.Now().UTC()),
		engine:                 engine,
		connString:             connString,
		connectionTimeout:      time.Duration(6) * time.Second,
		maxReconnectionRetries: 3,
		dieChan:                make(chan bool),
		topic:                  cluster.NewNatsSubject(subjectName, 1024),
		pending:                make(map[uint64]concepts.IMsgReq),
	}
	rpcClient.closed.Store(false)

	for _, opt := range opts {
		opt(rpcClient)
	}

	return rpcClient
}

func (rpc *RPCClient) Init() error {
	conn, err := cluster.SetupNatsConn(
		rpc.id,
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

func (rpc *RPCClient) Run() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("rpc have panic err:", r, string(debug.Stack()))
		}
		rpc.Stop()
	}()

	process := func(natsMsg *nats.Msg) {
		fmt.Printf("msg:%+v", natsMsg)
		fmt.Printf("data:%s", string(natsMsg.Data))

		if len(natsMsg.Data) == 0 {
			slog.Error("nats receive", "header", natsMsg.Header, "Subject", natsMsg.Subject, "Reply", natsMsg.Reply)
			return
		}
		response, err := msg.ResponseUnmarshal(natsMsg.Data)
		if err != nil {
			fmt.Printf("err:%+v", err)
			return
		}

		rpc.HandleResponse(rpc.seqId, response)
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

func (rpc *RPCClient) Send(topic string, data []byte) error {
	if rpc.closed.Load() {
		return errors.New("rpc client has closed")
	}

	reply := rpc.getReplySubject()
	return rpc.conn.PublishRequest(topic, reply, data)
}

func (rpc *RPCClient) SendRequest(request concepts.IMsgReq) error {
	if rpc.closed.Load() {
		return errors.New("rpc client has closed")
	}

	reply := rpc.getReplySubject()
	request.GetSender().Address = reply
	data, err := request.Marshal()
	if err != nil {
		return err
	}

	rpc.pendingMu.Lock()
	defer rpc.pendingMu.Unlock()
	rpc.seqId++
	request.SetSeqId(rpc.seqId)
	rpc.pending[rpc.seqId] = request

	return rpc.conn.PublishRequest(request.GetTarget().Address, reply, data)
}

func (rpc *RPCClient) HandleResponse(id uint64, resp concepts.IMsgResp) error {
	if rpc.closed.Load() {
		return errors.New("rpc client has closed")
	}

	rpc.pendingMu.Lock()
	defer rpc.pendingMu.Unlock()
	call, ok := rpc.pending[id]
	if !ok {
		return errors.New("invalid id")
	}
	delete(rpc.pending, id)

	call.HandleResponse(resp)
	return nil
}

func (rpc *RPCClient) Stop() {
	if rpc.closed.Load() {
		return
	}
	rpc.closed.Store(true)
	rpc.topic.Stop()
	rpc.conn.Close()
}

func (rpc *RPCClient) getReplySubject() string {
	return rpc.topic.Subject
}

func (rpc *RPCClient) GetAddress() string {
	return rpc.topic.Subject
}
