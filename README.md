# RPC 示例
```
// Server
type ActorObjB struct {
	*actor.Actor
	inited atomic.Bool
	id     int
}

func (actor *ActorObjB) OnInit() error {
	if actor.inited.Load() {
		return errors.New("duplicate init")
	}
	actor.Register(1, actor.Func1)
	actor.inited.Store(true)
	return nil
}

func (actor *ActorObjB) OnShutdown() {

}

func (actor *ActorObjB) Func1(ctx context.Context, arg *common_msg.Person, reply *common_msg.Person) errs.CodeError {
	reply.Age += arg.Age
	reply.Name = "Func1 hello world"
	reply.Address = actor.ActorId().ID
	fmt.Printf("inside value:%v\n", reply)

	return nil
}


func TestServer1(t *testing.T) {
	engine := actor.NewEngine(0, 1, 1001, "nats://127.0.0.1:4222")
	engine.MustInit()
	actorObj1 := &ActorObjB{
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

type ActorObjA struct {
	*actor.Actor
	inited atomic.Bool
	id     int
}

func (actor *ActorObjA) OnInit() error {
	if actor.inited.Load() {
		return errors.New("duplicate init")
	}
	return nil
}

func (actor *ActorObjA) OnShutdown() {

}

func TestClient1(t *testing.T) {
	engine := actor.NewEngine(0, 1, 1002, "nats://127.0.0.1:4222")

	engine.MustInit()
	engine.Start()

	actorObj1 := &ActorObjA{
		Actor: actor.NewActor("1", engine),
	}

	echoObj := &rpc_msg.RPC_EchoTestRequest{Value1: 12345678, Value2: "小明"}

	//engine.0.1.1001.server
	echoResponse, err := actor.SendRequest[rpc_msg.RPC_EchoTestResponse](actorObj1, concepts.NewActorId("engine.0.1.1001.server", "1"), 1001, echoObj)
	if err != nil {
		fmt.Println("err", err)
	}
	fmt.Printf("\n\n\n================echoResponse:%T, %v\n", echoResponse, echoResponse)

	time.Sleep(time.Duration(30) * time.Second)
}
```