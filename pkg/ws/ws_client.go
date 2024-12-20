package ws

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"runtime/debug"
	"sync"
	"sync/atomic"
	"time"

	"github.com/wuqunyong/file_storage/pkg/actor"
	"github.com/wuqunyong/file_storage/pkg/common/concepts"
)

const (
	// MessageText is for UTF-8 encoded text messages like JSON.
	MessageText = iota + 1
	// MessageBinary is for binary messages like protobufs.
	MessageBinary
	// CloseMessage denotes a close control message. The optional message
	// payload contains a numeric code and text. Use the FormatCloseMessage
	// function to format a close message payload.
	CloseMessage = 8

	// PingMessage denotes a ping control message. The optional message payload
	// is UTF-8 encoded text.
	PingMessage = 9

	// PongMessage denotes a pong control message. The optional message payload
	// is UTF-8 encoded text.
	PongMessage = 10
)

var (
	ErrConnClosed                = errors.New("conn has closed")
	ErrNotSupportMessageProtocol = errors.New("not support message protocol")
	ErrClientClosed              = errors.New("client actively close the connection")
	ErrPanic                     = errors.New("panic error")
	ErrDataUnmarshal             = errors.New("data Unmarshal error")
	ErrUnregisterOpcode          = errors.New("unregister opcode")

	connId uint64 = 0
)

func IncrementConnId() uint64 {
	return atomic.AddUint64(&connId, 1)
}

type UserConnContext struct {
	ConnID     uint64
	RespWriter http.ResponseWriter
	Req        *http.Request
	Path       string
	Method     string
	RemoteAddr string
}

func newContext(respWriter http.ResponseWriter, req *http.Request) *UserConnContext {
	return &UserConnContext{
		ConnID:     IncrementConnId(),
		RespWriter: respWriter,
		Req:        req,
		Path:       req.URL.Path,
		Method:     req.Method,
		RemoteAddr: req.RemoteAddr,
	}
}

type Client struct {
	UserID         string `json:"userID"`
	w              *sync.Mutex
	conn           LongConn
	ctx            *UserConnContext
	longConnServer LongConnServer
	closed         atomic.Bool
	closedErr      error
	hbCtx          context.Context
	hbCancel       context.CancelFunc
	concepts.IActor
}

func (c *Client) ResetClient(ctx *UserConnContext, conn LongConn, longConnServer LongConnServer, id string) {
	c.w = new(sync.Mutex)
	c.conn = conn
	c.ctx = ctx
	c.longConnServer = longConnServer
	c.closed.Store(false)
	c.closedErr = nil
	c.hbCtx, c.hbCancel = context.WithCancel(context.Background())
	c.IActor = actor.NewActor(id, longConnServer.GetEngine())
}

func (c *Client) Run() {
	go c.readMessage()
	go c.longConnServer.GetEngine().SpawnActor(c.IActor)
}

func (c *Client) GetActor() concepts.IActor {
	return c.IActor
}

func (c *Client) readMessage() {
	defer func() {
		if r := recover(); r != nil {
			c.closedErr = ErrPanic
			fmt.Println("socket have panic err:", r, string(debug.Stack()))
		}
		c.close()
	}()

	c.activeHeartbeat()

	for {
		messageType, message, returnErr := c.conn.ReadMessage()
		if returnErr != nil {
			if netErr, ok := returnErr.(net.Error); ok && netErr.Timeout() {
				fmt.Println("Read timeout, closing connection...")
			} else {
				fmt.Println("Read error:", returnErr)
			}
			c.closedErr = returnErr
			return
		}

		if c.closed.Load() {
			// The scenario where the connection has just been closed, but the coroutine has not exited
			c.closedErr = ErrConnClosed
			return
		}

		switch messageType {
		case MessageBinary:
		case MessageText:
			fmt.Printf("message:%s\n", message)

			var jsonReq Req
			err := json.Unmarshal(message, &jsonReq)
			if err != nil {
				fmt.Printf("Unmarshal err: %s\n", err.Error())
				c.closedErr = ErrDataUnmarshal
				return
			}

			fmt.Printf("Unmarshal jsonReq: %+v\n", jsonReq)

			handler := GetInstance().GetHandler(jsonReq.Opcode)
			if handler == nil {
				fmt.Printf("unregister opcode:%d\n", jsonReq.Opcode)
				c.closedErr = ErrUnregisterOpcode
				return
			}

			task := func() {
				resp, codeErr := handler(c, &jsonReq)
				if codeErr != nil {
					reply := Resp{
						RequestId: jsonReq.RequestId,
						ErrCode:   codeErr.Code(),
						ErrMsg:    codeErr.Error(),
					}
					c.writeTextMsg(reply)
				} else {
					reply := Resp{
						RequestId: jsonReq.RequestId,
						Data:      resp,
					}
					c.writeTextMsg(reply)
				}
			}
			c.GetActor().PostTask(task)

		case PingMessage:
			c.writePongMsg("")

		case CloseMessage:
			c.closedErr = ErrClientClosed
			return

		default:
		}
	}
}

func (c *Client) close() {
	c.w.Lock()
	defer c.w.Unlock()
	if c.closed.Load() {
		return
	}
	c.closed.Store(true)
	c.conn.Close()
	c.hbCancel() // Close server-initiated heartbeat.
	c.longConnServer.UnRegister(c)
}

func (c *Client) activeHeartbeat() {
	go func() {
		fmt.Printf("server initiative send heartbeat start.\n")
		ticker := time.NewTicker(pingPeriod)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if err := c.writePingMsg(); err != nil {
					fmt.Printf("send Ping Message error.\n")
					return
				}
			case <-c.hbCtx.Done():
				return
			}
		}
	}()

}

func (c *Client) writePingMsg() error {
	if c.closed.Load() {
		return nil
	}

	c.w.Lock()
	defer c.w.Unlock()

	err := c.conn.SetWriteDeadline(writeWait)
	if err != nil {
		return err
	}

	return c.conn.WriteMessage(PingMessage, nil)
}

func (c *Client) writePongMsg(appData string) error {
	if c.closed.Load() {
		return nil
	}

	c.w.Lock()
	defer c.w.Unlock()

	err := c.conn.SetWriteDeadline(writeWait)
	if err != nil {
		return err
	}
	err = c.conn.WriteMessage(PongMessage, []byte(appData))
	return err
}

func (c *Client) writeBinaryMsg(resp Resp) error {
	if c.closed.Load() {
		return nil
	}

	encodedBuf, err := c.longConnServer.Encode(resp)
	if err != nil {
		return err
	}

	c.w.Lock()
	defer c.w.Unlock()

	err = c.conn.SetWriteDeadline(writeWait)
	if err != nil {
		return err
	}

	return c.conn.WriteMessage(MessageBinary, encodedBuf)
}

func (c *Client) writeTextMsg(resp Resp) error {
	if c.closed.Load() {
		return nil
	}

	encodedBuf, err := c.longConnServer.Encode(resp)
	if err != nil {
		return err
	}

	c.w.Lock()
	defer c.w.Unlock()

	err = c.conn.SetWriteDeadline(writeWait)
	if err != nil {
		return err
	}

	return c.conn.WriteMessage(MessageText, encodedBuf)
}
