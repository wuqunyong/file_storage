package concepts

type IRPCClient interface {
	Init() error
	Run()
	Send(topic string, data []byte) error
	SendRequest(request IMsgReq) error
	HandleResponse(id int64, resp IMsgResp) error
	Stop()
}

type IRPCServer interface {
	Init() error
	Run()
	HandleRequest(request IMsgReq) error
	SendResponse(subj string, response IMsgResp) error
	Stop()
}
