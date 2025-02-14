package easytcp

import (
	"fmt"
	"net"
	"time"
)

type Client struct {
	// Packer is the message packer, will be passed to session.
	Packer Packer

	// Codec is the message codec, will be passed to session.
	Codec Codec

	socketReadBufferSize  int
	socketWriteBufferSize int
	readTimeout           time.Duration
	writeTimeout          time.Duration
	respQueueSize         int
	router                *Router
	printRoutes           bool
	stoppedC              chan struct{}
	asyncRouter           bool
}

type ClientOption struct {
	SocketReadBufferSize  int           // sets the socket read buffer size.
	SocketWriteBufferSize int           // sets the socket write buffer size.
	ReadTimeout           time.Duration // sets the timeout for connection read.
	WriteTimeout          time.Duration // sets the timeout for connection write.
	Packer                Packer        // packs and unpacks packet payload, default packer is the DefaultPacker.
	Codec                 Codec         // encodes and decodes the message data, can be nil.
	RespQueueSize         int           // sets the response channel size of session, DefaultRespQueueSize will be used if < 0.
	DoNotPrintRoutes      bool          // whether to print registered route handlers to the console.

	// AsyncRouter represents whether to execute a route HandlerFunc of each session in a goroutine.
	// true means execute in a goroutine.
	AsyncRouter bool
}

var ErrClientStopped = fmt.Errorf("client stopped")

func NewClient(opt *ClientOption) *Client {
	if opt.Packer == nil {
		opt.Packer = NewDefaultPacker()
	}
	if opt.RespQueueSize < 0 {
		opt.RespQueueSize = DefaultRespQueueSize
	}
	return &Client{
		socketReadBufferSize:  opt.SocketReadBufferSize,
		socketWriteBufferSize: opt.SocketWriteBufferSize,
		respQueueSize:         opt.RespQueueSize,
		readTimeout:           opt.ReadTimeout,
		writeTimeout:          opt.WriteTimeout,
		Packer:                opt.Packer,
		Codec:                 opt.Codec,
		printRoutes:           !opt.DoNotPrintRoutes,
		router:                newRouter(),
		stoppedC:              make(chan struct{}),
		asyncRouter:           opt.AsyncRouter,
	}
}

func (c *Client) Dial(address string) (*session, error) {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return nil, err
	}

	sess := newSession(conn, &sessionOption{
		Packer:        c.Packer,
		Codec:         c.Codec,
		respQueueSize: c.respQueueSize,
		asyncRouter:   c.asyncRouter,
	})
	close(sess.afterCreateHookC)

	go c.handleConn(conn, sess)
	return sess, nil
}

// handleConn creates a new session with `conn`,
// handles the message through the session in different goroutines,
// and waits until the session's closed, then close the `conn`.
func (c *Client) handleConn(conn net.Conn, sess *session) {
	defer conn.Close() // nolint

	go sess.readInbound(c.router, c.readTimeout) // start reading message packet from connection.
	go sess.writeOutbound(c.writeTimeout)        // start writing message packet to connection.

	select {
	case <-sess.closedC: // wait for session finished.
	case <-c.stoppedC: // or the server is stopped.
	}

	close(sess.afterCloseHookC)
}

func (c *Client) Stop() error {
	close(c.stoppedC)
	return nil
}

// AddRoute registers message handler and middlewares to the router.
func (c *Client) AddRoute(msgID interface{}, handler HandlerFunc, middlewares ...MiddlewareFunc) {
	c.router.register(msgID, handler, middlewares...)
}

// Use registers global middlewares to the router.
func (c *Client) Use(middlewares ...MiddlewareFunc) {
	c.router.registerMiddleware(middlewares...)
}

// NotFoundHandler sets the not-found handler for router.
func (c *Client) NotFoundHandler(handler HandlerFunc) {
	c.router.setNotFoundHandler(handler)
}

func (c *Client) isStopped() bool {
	select {
	case <-c.stoppedC:
		return true
	default:
		return false
	}
}
