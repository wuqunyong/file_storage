# RPC 示例
```
// Server
type ActorService struct {
	*actor.Actor
	inited atomic.Bool
}

func (actor *ActorService) OnInit() error {
	if actor.inited.Load() {
		return errors.New("duplicate init")
	}
	actor.Register(1, actor.Func1)
	actor.Register(2, actor.Func2)
	actor.Register(1001, actor.EchoTest)
	actor.Register(1002, actor.NotifyTest)
	actor.inited.Store(true)
	return nil
}

func (actor *ActorService) OnShutdown() {
}

// 处理：请求-响应 操作，结果正常
func (actor *ActorService) Func1(ctx context.Context, request *common_msg.EchoRequest, response *common_msg.EchoResponse) errs.CodeError {
	logger.Log(logger.InfoLevel, "Func1", "request", request)

	response.Value1 = request.Value1 + 1
	response.Value2 = request.Value2 + " | response"
	return nil
}

// 处理：请求-响应 操作，结果异常
func (actor *ActorService) Func2(ctx context.Context, request *common_msg.EchoRequest, response *common_msg.EchoResponse) errs.CodeError {
	logger.Log(logger.InfoLevel, "Func2", "request", request)

	return errs.NewCodeError(errors.New("invalid"), 123)
}

// 处理：请求-响应 操作，结果正常
func (actor *ActorService) EchoTest(ctx context.Context, request *rpc_msg.RPC_EchoTestRequest, response *rpc_msg.RPC_EchoTestResponse) errs.CodeError {
	response.Value1 = request.Value1
	response.Value2 = request.Value2 + "| Response"

	logger.Log(logger.InfoLevel, "EchoTest", "request", request, "response", response)
	return nil
}

// 处理：通知 操作
func (actor *ActorService) NotifyTest(ctx context.Context, notify *rpc_msg.RPC_EchoTestRequest) {
	logger.Log(logger.InfoLevel, "NotifyTest", "notify", notify)
}

func TestServer1(t *testing.T) {
	//http://127.0.0.1:8222/connz?subs=true

	engine := actor.NewEngine(0, 1, 1001, "nats://127.0.0.1:4222")
	engine.MustInit()
	actorObj1 := &ActorService{
		Actor: actor.NewActor("1", engine),
	}
	engine.SpawnActor(actorObj1)
	defer engine.Stop()
	engine.Start()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs
}


// Client
type ActorClient struct {
	*actor.Actor
	inited atomic.Bool
}

func (actor *ActorClient) OnInit() error {
	if actor.inited.Load() {
		return errors.New("duplicate init")
	}
	actor.inited.Store(true)
	return nil
}

func (actor *ActorClient) OnShutdown() {

}

func TestClient1(t *testing.T) {
	engine := actor.NewEngine(0, 1, 1002, "nats://127.0.0.1:4222")

	engine.MustInit()
	engine.Start()

	actorObj1 := &ActorClient{
		Actor: actor.NewActor("1", engine),
	}

	echo := &common_msg.EchoRequest{Value1: 123456, Value2: "小明"}
	obj1, err := actor.SendRequest[common_msg.EchoResponse](actorObj1, concepts.NewActorId("engine.0.1.1001.server", "1"), 1, echo)
	if err != nil {
		logger.Log(logger.InfoLevel, "opcode 1 failure", "request", echo, "response", obj1, "err", err)
	} else {
		logger.Log(logger.InfoLevel, "opcode 1 success", "request", echo, "response", obj1)
	}

	obj2, err := actor.SendRequest[common_msg.EchoResponse](actorObj1, concepts.NewActorId("engine.0.1.1001.server", "1"), 2, echo)
	if err != nil {
		logger.Log(logger.InfoLevel, "opcode 2 failure", "request", echo, "response", obj2, "err", err)
	} else {
		logger.Log(logger.InfoLevel, "opcode 2 success", "request", echo, "response", obj2)
	}

	echoObj := &rpc_msg.RPC_EchoTestRequest{Value1: 12345678, Value2: "小明"}
	echoResponse, err := actor.SendRequest[rpc_msg.RPC_EchoTestResponse](actorObj1, concepts.NewActorId("engine.0.1.1001.server", "1"), 1001, echoObj)
	if err != nil {
		logger.Log(logger.InfoLevel, "opcode 1001 failure", "request", echoObj, "response", echoResponse, "err", err)
	} else {
		logger.Log(logger.InfoLevel, "opcode 1001 success", "request", echoObj, "response", echoResponse)
	}

	sendErr := actor.SendNotify(actorObj1, concepts.NewActorId("engine.0.1.1001.server", "1"), 1002, echoObj)
	if sendErr != nil {
		logger.Log(logger.InfoLevel, "opcode 1002 failure", "notify", echoObj, "err", sendErr)
	}

	time.Sleep(time.Duration(1800) * time.Second)
}
```