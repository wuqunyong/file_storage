package tcpserver

import (
	"fmt"
	"log/slog"

	"github.com/wuqunyong/file_storage/pkg/common/concepts"
	"github.com/wuqunyong/file_storage/pkg/easytcp"
	testdata "github.com/wuqunyong/file_storage/protobuf"
)

type TCPServer struct {
	server  *easytcp.Server
	engine  concepts.IEngine
	address string
}

func NewTCPServer(opt *easytcp.ServerOption, address string) *TCPServer {
	server := easytcp.NewServer(opt)
	return &TCPServer{
		server:  server,
		address: address,
	}
}

func (s *TCPServer) Name() string {
	return "tcpserver"
}

func (s *TCPServer) Priority() int32 {
	return 1
}

func (s *TCPServer) SetEngine(engine concepts.IEngine) {
	s.engine = engine
}

func (s *TCPServer) GetEngine() concepts.IEngine {
	return s.engine
}

func (s *TCPServer) OnInit() error {
	s.server.AddRoute(1001, func(c easytcp.Context) {
		var reqData testdata.AccountLoginRequest
		err := c.Bind(&reqData)
		if err != nil {
			fmt.Println("err:", err)
		}
		slog.Info("Recv 1001", "data", reqData.String())

		// set response
		//c.SetResponseMessage(easytcp.NewMessage(1002, []byte("copy that")))
	})
	s.server.AddRoute(1103, func(c easytcp.Context) {
		var reqData testdata.EchoRequest
		err := c.Bind(&reqData)
		if err != nil {
			fmt.Println("err:", err)
		}
		slog.Info("Recv 1003", "data", reqData.String())

		// set response
		var response testdata.EchoResponse
		response.Value1 = reqData.Value1
		response.Value2 = reqData.Value2 + "|response"
		c.SetResponse(1104, &response)
	})
	return nil
}

func (s *TCPServer) OnStart() {
	go func() {
		var err error
		if err = s.server.Run(s.address); err != nil && err != easytcp.ErrServerStopped {
			panic(fmt.Sprintf("serve error:%s", err.Error()))
		}
		slog.Info("Stopped tcpserver", "err", err)
	}()
}

func (s *TCPServer) OnCleanup() {
	s.server.Stop()
}

func NewPBServerOption() *easytcp.ServerOption {
	packer := NewPBPacker()
	codec := &easytcp.ProtobufCodec{}
	return &easytcp.ServerOption{Packer: packer,
		Codec: codec}
}
