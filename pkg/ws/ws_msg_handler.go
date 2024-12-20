package ws

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"sync"
	"time"

	"github.com/wuqunyong/file_storage/pkg/encoders"
	"github.com/wuqunyong/file_storage/pkg/errs"
	"github.com/wuqunyong/file_storage/pkg/funcutils"
	"github.com/wuqunyong/file_storage/pkg/tick"
	testdata "github.com/wuqunyong/file_storage/protobuf"
	"golang.org/x/exp/rand"
	"google.golang.org/protobuf/proto"
)

type Req struct {
	RequestId int32  `json:"requestId"`
	Opcode    int32  `json:"opcode"`
	Flag      int32  `json:"flag"`
	Data      []byte `json:"data"` // 在序列化和反序列化时，[]byte 会被自动转换为 Base64 编码的字符串（JSON 格式要求）
}

type Resp struct {
	RequestId int32  `json:"requestId"`
	ErrCode   int32  `json:"errCode"`
	ErrMsg    string `json:"errMsg"`
	Data      []byte `json:"data"`
}

func NewResp(requestId int32, errCode int32, errMsg string) *Resp {
	return &Resp{
		RequestId: requestId,
		ErrCode:   errCode,
		ErrMsg:    errMsg,
	}
}

var (
	instance *RegisterHandler
	once     sync.Once
)

type RegisterHandler struct {
	msgHandler map[int32]*funcutils.MethodType
}

func newRegisterHandler() *RegisterHandler {
	return &RegisterHandler{
		msgHandler: make(map[int32]*funcutils.MethodType),
	}
}

func GetInstance() *RegisterHandler {
	once.Do(func() {
		instance = newRegisterHandler()
	})
	return instance
}

func (h *RegisterHandler) Register(opcode int32, handler any) error {
	_, ok := h.msgHandler[opcode]
	if ok {
		sErr := fmt.Sprintf("duplicate id:%d\n", opcode)
		return errors.New(sErr)
	}
	ptrMethon, err := funcutils.GetClientReflectFunc(handler)
	if err != nil {
		return err
	}
	h.msgHandler[opcode] = ptrMethon
	return nil
}

func (h *RegisterHandler) GetHandler(opcode int32) *funcutils.MethodType {
	handler, ok := h.msgHandler[opcode]
	if ok {
		return handler
	}

	return nil
}

func (handler *RegisterHandler) callFunc(client *Client, request *Req) *Resp {
	requestId := request.RequestId
	ptrMethod := handler.GetHandler(request.Opcode)
	if ptrMethod == nil {
		sError := fmt.Sprintf("unregister Opcode:%d", request.Opcode)
		reply := NewResp(requestId, 1, sError)
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
		sError := fmt.Sprintf("Decode err:%v\n", err)
		reply := NewResp(requestId, 1, sError)
		return reply
	}
	code, err = funcutils.CallClientReflectFunc(ptrMethod, client, args, response)
	if err != nil {
		sError := fmt.Sprintf("err:%s\n" + err.Error())
		reply := NewResp(requestId, 1, sError)
		return reply
	}

	if code != nil {
		reply := NewResp(requestId, code.Code(), code.Msg())
		return reply
	}

	replyData, err := decoder.Encode(response)
	if err != nil {
		sError := fmt.Sprintf("Encode err:%v\n", err)
		reply := NewResp(requestId, 1, sError)
		return reply
	}

	reply := NewResp(requestId, 0, "")
	reply.Data = replyData
	return reply
}

type ModuleA struct {
}

type TestJsonMsg struct {
	Hello string `json:"hello"`
	Value int    `json:"value"`
	Extra string `json:"extra"`
}

func (a *ModuleA) Handler_Func1(client *Client, data *Req) ([]byte, errs.CodeError) {
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

func (a *ModuleA) Handler_Func2(client *Client, data *Req) ([]byte, errs.CodeError) {
	//{"opcode":2,"data":[10,5,100,101,114,101,107,16,22,26,21,49,52,48,32,78,101,119,32,77,111,110,116,103,111,109,101,114,121,32,83,116,82,37,10,3,115,97,109,18,30,10,3,115,97,109,16,19,26,21,49,52,48,32,78,101,119,32,77,111,110,116,103,111,109,101,114,121,32,83,116,82,37,10,3,109,101,103,18,30,10,3,109,101,103,16,17,26,21,49,52,48,32,78,101,119,32,77,111,110,116,103,111,109,101,114,121,32,83,116]}

	person := &testdata.Person{}
	err := proto.Unmarshal(data.Data, person)
	if err != nil {
		return nil, errs.NewCodeError(err, errs.CODE_Unmarshal)
	}

	fmt.Printf("other:%+v\n", person)

	person.Address += "|Changed"

	respBytes, err := proto.Marshal(person)
	if err != nil {
		return nil, errs.NewCodeError(err, errs.CODE_Marshal)
	}

	curMilliTime := time.Now().UnixMilli()
	for i := 0; i < 60; i++ {
		iRand := rand.Intn(180) * 1000
		iValue := curMilliTime + int64(iRand)
		item := tick.NewItem(iValue, func(id uint64) {
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

func (a *ModuleA) Handler_Func3(client *Client, reqeust *testdata.Person, response *testdata.Person) errs.CodeError {
	//{"opcode":2,"data":[10,5,100,101,114,101,107,16,22,26,21,49,52,48,32,78,101,119,32,77,111,110,116,103,111,109,101,114,121,32,83,116,82,37,10,3,115,97,109,18,30,10,3,115,97,109,16,19,26,21,49,52,48,32,78,101,119,32,77,111,110,116,103,111,109,101,114,121,32,83,116,82,37,10,3,109,101,103,18,30,10,3,109,101,103,16,17,26,21,49,52,48,32,78,101,119,32,77,111,110,116,103,111,109,101,114,121,32,83,116]}

	fmt.Printf("recv:%+v\n", reqeust)

	response.Address += "|Changed"
	curMilliTime := time.Now().UnixMilli()
	for i := 0; i < 60; i++ {
		iRand := rand.Intn(180) * 1000
		iValue := curMilliTime + int64(iRand)
		item := tick.NewItem(iValue, func(id uint64) {
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
