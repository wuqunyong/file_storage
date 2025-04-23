package ws

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"sync"
	"time"

	"github.com/wuqunyong/file_storage/pkg/common"
	"github.com/wuqunyong/file_storage/pkg/encoders"
	"github.com/wuqunyong/file_storage/pkg/errs"
	"github.com/wuqunyong/file_storage/pkg/funcutils"
	"github.com/wuqunyong/file_storage/pkg/tick"
	"github.com/wuqunyong/file_storage/proto/common_msg"
	"golang.org/x/exp/rand"
	"google.golang.org/protobuf/proto"
)

var (
	instance *RegisterHandler
	once     sync.Once
)

type RegisterHandler struct {
	msgHandler map[int32]*funcutils.MethodType
}

func NewRegisterHandler() *RegisterHandler {
	return &RegisterHandler{
		msgHandler: make(map[int32]*funcutils.MethodType),
	}
}

func GetInstance() *RegisterHandler {
	once.Do(func() {
		instance = NewRegisterHandler()
	})
	return instance
}

func (h *RegisterHandler) MustRegister(opcode int32, handler any) {
	_, ok := h.msgHandler[opcode]
	if ok {
		sErr := fmt.Sprintf("duplicate id:%d", opcode)
		panic(sErr)
	}
	ptrMethon, err := funcutils.GetClientReflectFunc(handler)
	if err != nil {
		panic(err)
	}
	h.msgHandler[opcode] = ptrMethon
}

func (h *RegisterHandler) GetHandler(opcode int32) *funcutils.MethodType {
	handler, ok := h.msgHandler[opcode]
	if ok {
		return handler
	}

	return nil
}

func (handler *RegisterHandler) CallFunc(client *Client, request *common.Req) *common.Resp {
	requestId := request.RequestId
	ptrMethod := handler.GetHandler(request.Opcode)
	if ptrMethod == nil {
		sError := fmt.Sprintf("unregister Opcode:%d", request.Opcode)
		reply := common.NewResp(requestId, 1, sError)
		return reply
	}

	var (
		code     errs.CodeError
		err      error
		decoder  = encoders.NewProtobufEncoder()
		response = reflect.New(ptrMethod.ReplyType.Elem()).Interface()
	)

	args := reflect.New(ptrMethod.ArgType[1].Elem()).Interface()
	err = decoder.Decode(request.Data, args)
	if err != nil {
		sError := fmt.Sprintf("Decode err:%v", err)
		reply := common.NewResp(requestId, 1, sError)
		return reply
	}
	code, err = funcutils.CallClientReflectFunc(ptrMethod, client, args, response)
	if err != nil {
		sError := fmt.Sprintf("err:%s" + err.Error())
		reply := common.NewResp(requestId, 1, sError)
		return reply
	}

	if code != nil {
		reply := common.NewResp(requestId, code.Code(), code.Msg())
		return reply
	}

	replyData, err := decoder.Encode(response)
	if err != nil {
		sError := fmt.Sprintf("Encode err:%v", err)
		reply := common.NewResp(requestId, 1, sError)
		return reply
	}

	reply := common.NewResp(requestId, 0, "")
	reply.Data = replyData
	return reply
}

type ClientHandler struct {
	client     *Client
	msgHandler *RegisterHandler
}

func NewClientHandler(client *Client, msgHandler *RegisterHandler) *ClientHandler {
	return &ClientHandler{
		client:     client,
		msgHandler: msgHandler,
	}
}

func (handler *ClientHandler) Init() (err error) {
	defer func() {
		if r := recover(); r != nil {
			sErr := fmt.Sprintf("register have panic err:%s", r)
			err = errors.New(sErr)
		}
	}()

	moduleA := &ModuleA{}
	handler.msgHandler.MustRegister(3, moduleA.Handler_Func3)
	handler.msgHandler.MustRegister(4, moduleA.Handler_Func4)
	// handler.msgHandler.MustRegister(4, moduleA.Handler_Func4)
	handler.msgHandler.MustRegister(5, moduleA.Handler_Func4)
	return err
}

func (handler *ClientHandler) CallFunc(request *common.Req) {
	response := handler.msgHandler.CallFunc(handler.client, request)
	handler.client.writeTextMsg(response)
}

type ModuleA struct {
}

type TestJsonMsg struct {
	Hello string `json:"hello"`
	Value int    `json:"value"`
	Extra string `json:"extra"`
}

func (a *ModuleA) Handler_Func1(client *Client, data *common.Req) ([]byte, errs.CodeError) {
	//{"hello":"world","value":123}
	//eyJoZWxsbyI6IndvcmxkIiwidmFsdWUiOjEyM30=
	//{"opcode":1,"data":"eyJoZWxsbyI6IndvcmxkIiwidmFsdWUiOjEyM30="}

	var obj TestJsonMsg
	if err := json.Unmarshal(data.Data, &obj); err != nil {
		return nil, errs.NewCodeError(err, errs.CODE_Unmarshal)
	}
	obj.Extra += "|Handler_Func1"

	respBytes, err := json.Marshal(obj)
	if err != nil {
		return nil, errs.NewCodeError(err, errs.CODE_Marshal)
	}
	return respBytes, nil
}

func (a *ModuleA) Handler_Func2(client *Client, data *common.Req) ([]byte, errs.CodeError) {
	//{"opcode":2,"data":[10,5,100,101,114,101,107,16,22,26,21,49,52,48,32,78,101,119,32,77,111,110,116,103,111,109,101,114,121,32,83,116,82,37,10,3,115,97,109,18,30,10,3,115,97,109,16,19,26,21,49,52,48,32,78,101,119,32,77,111,110,116,103,111,109,101,114,121,32,83,116,82,37,10,3,109,101,103,18,30,10,3,109,101,103,16,17,26,21,49,52,48,32,78,101,119,32,77,111,110,116,103,111,109,101,114,121,32,83,116]}

	person := &common_msg.Person{}
	err := proto.Unmarshal(data.Data, person)
	if err != nil {
		return nil, errs.NewCodeError(err, errs.CODE_Unmarshal)
	}

	fmt.Printf("other:%+v", person)

	person.Address += "|Changed"

	respBytes, err := proto.Marshal(person)
	if err != nil {
		return nil, errs.NewCodeError(err, errs.CODE_Marshal)
	}

	// curMilliTime := time.Now().UnixMilli()
	for i := 0; i < 60; i++ {
		iRand := rand.Intn(180) * 1000
		iValue := time.Duration(iRand)
		item := tick.NewTimer(iValue*time.Millisecond, func(id uint64) {
			fmt.Println("Id:", id, "expireTime:", iValue)
		})
		client.GetTimerQueue().Push(item)
		fmt.Println("id:", item.GetId(), "set expireTime:", iValue)

		iMod := item.GetId() % 10
		if iMod%3 == 0 {
			client.GetTimerQueue().Remove(item.GetId())
			fmt.Println("remove id:", item.GetId())
		}
	}

	return respBytes, nil
}

func (a *ModuleA) Handler_Func3(client *Client, reqeust *common_msg.Person, response *common_msg.Person) errs.CodeError {
	//{"opcode":2,"data":[10,5,100,101,114,101,107,16,22,26,21,49,52,48,32,78,101,119,32,77,111,110,116,103,111,109,101,114,121,32,83,116,82,37,10,3,115,97,109,18,30,10,3,115,97,109,16,19,26,21,49,52,48,32,78,101,119,32,77,111,110,116,103,111,109,101,114,121,32,83,116,82,37,10,3,109,101,103,18,30,10,3,109,101,103,16,17,26,21,49,52,48,32,78,101,119,32,77,111,110,116,103,111,109,101,114,121,32,83,116]}

	fmt.Printf("recv:%+v", reqeust)

	response.Address += "|Changed"
	// curMilliTime := time.Now().UnixMilli()
	for i := 0; i < 60; i++ {
		iRand := rand.Intn(180) * 1000
		iValue := time.Duration(iRand)
		item := tick.NewTimer(iValue*time.Millisecond, func(id uint64) {
			fmt.Println("Id:", id, "expireTime:", iValue)
		})
		client.GetTimerQueue().Push(item)
		fmt.Println("id:", item.GetId(), "set expireTime:", iValue)

		iMod := item.GetId() % 10
		if iMod%3 == 0 {
			client.GetTimerQueue().Remove(item.GetId())
			fmt.Println("remove id:", item.GetId())
		}
	}

	return nil
}

func (a *ModuleA) Handler_Func4(client *Client, reqeust *common_msg.Person, response *common_msg.Person) errs.CodeError {
	//{"opcode":2,"data":[10,5,100,101,114,101,107,16,22,26,21,49,52,48,32,78,101,119,32,77,111,110,116,103,111,109,101,114,121,32,83,116,82,37,10,3,115,97,109,18,30,10,3,115,97,109,16,19,26,21,49,52,48,32,78,101,119,32,77,111,110,116,103,111,109,101,114,121,32,83,116,82,37,10,3,109,101,103,18,30,10,3,109,101,103,16,17,26,21,49,52,48,32,78,101,119,32,77,111,110,116,103,111,109,101,114,121,32,83,116]}

	fmt.Printf("recv:%+v", reqeust)

	response.Address += "|Changed"
	// curMilliTime := time.Now().UnixMilli()
	for i := 0; i < 60; i++ {
		iRand := rand.Intn(180) * 1000
		iValue := time.Duration(iRand)
		item := tick.NewTimer(iValue*time.Millisecond, func(id uint64) {
			fmt.Println("Id:", id, "expireTime:", iValue)
		})
		client.GetTimerQueue().Push(item)
		fmt.Println("id:", item.GetId(), "set expireTime:", iValue)

		iMod := item.GetId() % 10
		if iMod%3 == 0 {
			client.GetTimerQueue().Remove(item.GetId())
			fmt.Println("remove id:", item.GetId())
		}
	}

	return errs.NewCodeError(errors.New("customer err"), 123)
}
