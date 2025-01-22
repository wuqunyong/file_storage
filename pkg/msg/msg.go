package msg

import (
	"context"
	"errors"
	"fmt"
	"reflect"

	"github.com/wuqunyong/file_storage/pkg/common/concepts"
	"github.com/wuqunyong/file_storage/pkg/constants"
	"github.com/wuqunyong/file_storage/pkg/encoders"
	"github.com/wuqunyong/file_storage/pkg/errs"
	testdata "github.com/wuqunyong/file_storage/protobuf"
)

type MsgReq struct {
	TargetId  *concepts.ActorId
	Remote    bool
	SeqId     int64
	FuncName  string
	Args      any
	ArgsData  []byte
	Done      chan *MsgResp
	Err       error
	NumCalls  int32
	Ctx       context.Context
	CtxCancel context.CancelFunc

	Sender    *concepts.ActorId
	Codec     encoders.IEncoder
	RPCServer concepts.IRPCServer
}

func NewMsgReq(target *concepts.ActorId, method string, args any, ctx context.Context, coder encoders.IEncoder) *MsgReq {
	var cancel context.CancelFunc
	if ctx == nil {
		ctx, cancel = context.WithTimeout(context.Background(), constants.DefaultTimeout)
	} else {
		_, ok := ctx.Deadline()
		if ok {
			ctx, cancel = context.WithCancel(ctx)
		} else {
			ctx, cancel = context.WithTimeout(ctx, constants.DefaultTimeout)
		}
	}

	req := &MsgReq{
		TargetId:  target,
		Remote:    false,
		FuncName:  method,
		Args:      args,
		Ctx:       ctx,
		CtxCancel: cancel,
		Done:      make(chan *MsgResp),
		Err:       nil,
		Codec:     coder,
	}
	return req
}

func (req *MsgReq) GetTarget() *concepts.ActorId {
	return req.TargetId
}

func (req *MsgReq) GetSender() *concepts.ActorId {
	return req.Sender
}

func (req *MsgReq) SetSeqId(value int64) {
	req.SeqId = value
}

func (req *MsgReq) SetRemote(value bool) error {
	if value {
		req.Remote = value
		data, err := req.Codec.Encode(req.Args)
		if err != nil {
			req.Err = err
			return err
		}

		req.ArgsData = data
	}
	return nil
}

func (req *MsgReq) Result() (*MsgResp, error) {
	defer func() {
		req.CtxCancel()
	}()

	select {
	case resp := <-req.Done:
		return resp, nil
	case <-req.Ctx.Done():
		return nil, req.Ctx.Err()
	}
}

func (req *MsgReq) Send(resp concepts.IMsgResp) {
	if req.Remote {
		req.RPCServer.SendResponse(req.Sender.Address, resp)
		return
	}

	response, ok := resp.(*MsgResp)
	if !ok {
		return
	}
	req.Done <- response
}

func (req *MsgReq) HandleResponse(resp concepts.IMsgResp) {
	response, ok := resp.(*MsgResp)
	if !ok {
		return
	}
	req.Done <- response
}

func (req *MsgReq) Marshal() ([]byte, error) {
	encoder := encoders.NewProtobufEncoder()
	request := &testdata.RPC_REQUEST{}
	request.Client = &testdata.CLIENT_IDENTIFIER{
		Stub: &testdata.CHANNEL{
			Address: req.Sender.Address,
			Id:      req.Sender.ID,
		},
		SeqId: req.SeqId,
	}
	request.FuncName = req.FuncName
	request.ArgsData = req.ArgsData
	request.Server = &testdata.SERVER_IDENTIFIER{
		Stub: &testdata.CHANNEL{
			Address: req.TargetId.Address,
			Id:      req.TargetId.ID,
		},
	}
	return encoder.Encode(request)
}

func RequestUnmarshal(data []byte) (*MsgReq, error) {
	encoder := encoders.NewProtobufEncoder()

	rpcRequest := &testdata.RPC_REQUEST{}
	err := encoder.Decode(data, rpcRequest)
	if err != nil {
		return nil, err
	}
	request := NewMsgReq(concepts.NewActorId(rpcRequest.Server.Stub.Address, rpcRequest.Server.Stub.Id), rpcRequest.FuncName, nil, nil, encoders.NewProtobufEncoder())
	request.Remote = true
	request.SeqId = rpcRequest.GetClient().SeqId
	request.ArgsData = rpcRequest.ArgsData
	request.Sender = concepts.NewActorId(rpcRequest.Client.Stub.Address, rpcRequest.Client.Stub.Id)
	return request, nil
}

func ResponseUnmarshal(data []byte) (*MsgResp, error) {
	encoder := encoders.NewProtobufEncoder()

	response := &testdata.RPC_RESPONSE{}
	err := encoder.Decode(data, response)
	if err != nil {
		return nil, err
	}
	resp := NewMsgResp(response.Client.SeqId, response.Status.Code, response.Status.Msg, encoder)
	resp.ReplyData = response.ResultData
	return resp, nil
}

type MsgResp struct {
	Remote    bool
	SeqId     int64
	ErrCode   int32
	ErrMsg    string
	Reply     any
	ReplyData []byte

	Codec encoders.IEncoder
}

func NewMsgResp(seqId int64, errCode int32, errMsg string, codec encoders.IEncoder) *MsgResp {
	return &MsgResp{
		SeqId:   seqId,
		ErrCode: errCode,
		ErrMsg:  errMsg,
		Codec:   codec,
	}
}

func (resp *MsgResp) Marshal() ([]byte, error) {
	encoder := encoders.NewProtobufEncoder()
	response := &testdata.RPC_RESPONSE{}
	response.Client = &testdata.CLIENT_IDENTIFIER{
		SeqId: resp.SeqId,
	}
	response.Status = &testdata.STATUS{
		Code: resp.ErrCode,
		Msg:  resp.ErrMsg,
	}
	response.ResultData = resp.ReplyData
	return encoder.Encode(response)
}

func GetResult[T any](req concepts.IMsgReq) (result *T, code errs.CodeError) {
	params := reflect.ValueOf(result)
	if params.Kind() != reflect.Ptr {
		return nil, errs.NewCodeError(errors.New("template type invalid"))
	}

	argType := params.Type().Elem()
	if argType.Kind() == reflect.Ptr {
		return nil, errs.NewCodeError(errors.New("template type invalid"))
	}

	request, ok := req.(*MsgReq)
	if !ok {
		sError := fmt.Sprintf("type assertion error,%T", req)
		return nil, errs.NewCodeError(errors.New(sError))
	}

	if request.Err != nil {
		return nil, errs.NewCodeError(request.Err)
	}

	response, err := request.Result()
	if err != nil {
		return nil, errs.NewCodeError(err)
	}

	if response.ErrCode != 0 {
		return nil, errs.NewCodeError(errors.New(response.ErrMsg), response.ErrCode)
	}

	if request.Remote {
		var obj T
		err = response.Codec.Decode(response.ReplyData, &obj)
		if err != nil {
			return nil, errs.NewCodeError(err)
		}

		return &obj, nil
	}

	obj, ok := response.Reply.(*T)
	if !ok {
		sError := fmt.Sprintf("type assertion error,%T:%T", result, response.Reply)
		return nil, errs.NewCodeError(errors.New(sError))
	}
	return obj, nil
}
