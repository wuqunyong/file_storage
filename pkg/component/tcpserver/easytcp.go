package tcpserver

import (
	"fmt"

	"github.com/wuqunyong/file_storage/pkg/concepts"
	"github.com/wuqunyong/file_storage/pkg/easytcp"
	"github.com/wuqunyong/file_storage/pkg/logger"
	"github.com/wuqunyong/file_storage/proto/common_msg"
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
		var reqData common_msg.AccountLoginRequest
		err := c.Bind(&reqData)
		if err != nil {
			fmt.Println("err:", err)
		}
		logger.Log(logger.InfoLevel, "Recv 1001", "data", reqData.String())

		// set response
		//c.SetResponseMessage(easytcp.NewMessage(1002, []byte("copy that")))
	})
	s.server.AddRoute(1103, func(c easytcp.Context) {
		var reqData common_msg.EchoRequest
		err := c.Bind(&reqData)
		if err != nil {
			fmt.Println("err:", err)
		}
		logger.Log(logger.InfoLevel, "Recv 1003", "data", reqData.String())

		// set response
		var response common_msg.EchoResponse
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
		logger.Log(logger.InfoLevel, "Stopped tcpserver", "err", err)
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
