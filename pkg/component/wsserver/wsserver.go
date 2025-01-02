package wsserver

import (
	"github.com/wuqunyong/file_storage/pkg/common/concepts"
	"github.com/wuqunyong/file_storage/pkg/ws"
)

type WSServer struct {
	server *ws.WsServer
	engine concepts.IEngine
}

func NewWSServer(config ws.Config) *WSServer {
	server := ws.NewWsServer(config, nil)
	return &WSServer{
		server: server,
	}
}

func (s *WSServer) Name() string {
	return "wsserver"
}

func (s *WSServer) Priority() int32 {
	return 1
}

func (s *WSServer) SetEngine(engine concepts.IEngine) {
	s.engine = engine
	s.server.SetEngine(engine)
}

func (s *WSServer) GetEngine() concepts.IEngine {
	return s.engine
}

func (s *WSServer) OnInit() error {
	return nil
}

func (s *WSServer) OnStart() {
	s.server.Run()
}

func (s *WSServer) OnCleanup() {
	s.server.Stop()
}
