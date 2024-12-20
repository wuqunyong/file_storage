package ws

import (
	"log"
	"net"
	"net/http"
	"sync"
)

type LongConnServer interface {
	Run() error
	Register(c *Client)
	UnRegister(c *Client)
	GetClient() *Client

	Encoder
}

type WsServer struct {
	config     Config
	httpServer *http.Server
	wg         *sync.WaitGroup

	registerChan   chan *Client
	unregisterChan chan *Client
	clientPool     sync.Pool

	Encoder
}

func NewWsServer(config Config) *WsServer {
	return &WsServer{
		config:         config,
		wg:             &sync.WaitGroup{},
		registerChan:   make(chan *Client, 1000),
		unregisterChan: make(chan *Client, 1000),
		clientPool: sync.Pool{
			New: func() any {
				return new(Client)
			},
		},
		Encoder: NewJsonEncoder(),
	}
}

func (ws *WsServer) Run() error {
	var (
		client *Client
	)

	if ws.config.ServerCertificate != "" && ws.config.ServerPrivateKey != "" {
		go ws.runHTTPS()
	} else {
		go ws.runHTTP()
	}

	go func() {
		for {
			select {
			case client = <-ws.registerChan:
				ws.registerClient(client)
			case client = <-ws.unregisterChan:
				ws.unregisterClient(client)
			}
		}
	}()
	return nil
}

func (ws *WsServer) Register(c *Client) {
	ws.registerChan <- c
}

func (ws *WsServer) UnRegister(c *Client) {
	ws.unregisterChan <- c
}

func (ws *WsServer) GetClient() *Client {
	client := ws.clientPool.Get().(*Client)
	return client
}

func (ws *WsServer) registerClient(client *Client) {

}

func (ws *WsServer) unregisterClient(client *Client) {
}

func (ws *WsServer) startHTTP() {
	defer func() {
		ws.wg.Done()
		log.Printf("Stopped HTTP server")
	}()

	listener, err := net.Listen("tcp", ws.config.HttpPort)
	if err != nil {
		log.Fatalf("Failed to create HTTP server: %v", err)
	}

	router := newGinRouter(ws)
	ws.httpServer = &http.Server{Handler: router}

	log.Printf("Starting HTTP server at %s", listener.Addr())
	err = ws.httpServer.Serve(listener)
	if err != nil && err != http.ErrServerClosed {
		log.Fatalf("Failed to start HTTP server: %v", err)
	}
}

func (ws *WsServer) startHTTPS() {
	defer func() {
		ws.wg.Done()
		log.Printf("Stopped HTTPS server")
	}()

	listener, err := net.Listen("tcp", ws.config.HttpsPort)
	if err != nil {
		log.Fatalf("Failed to create HTTPS server: %v", err)
	}

	router := newGinRouter(ws)
	ws.httpServer = &http.Server{Handler: router}

	log.Printf("Starting HTTPS server on port %s", ws.config.HttpsPort)
	err = ws.httpServer.ServeTLS(listener, ws.config.ServerCertificate, ws.config.ServerPrivateKey)
	if err != nil && err != http.ErrServerClosed {
		log.Fatalf("Failed to start HTTPS server: %v", err)
	}
}

func (ws *WsServer) runHTTP() {
	ws.wg.Add(1)
	go ws.startHTTP()
	ws.wg.Wait()
}

func (ws *WsServer) runHTTPS() {
	ws.wg.Add(1)
	go ws.startHTTPS()
	ws.wg.Wait()
}

func (ws *WsServer) Stop() {
	ws.httpServer.Close()
}
