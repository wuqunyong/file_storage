package msg

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/wuqunyong/file_storage/nats_msg"
	"github.com/wuqunyong/file_storage/pkg/concepts"
	"github.com/wuqunyong/file_storage/pkg/constants"
	"github.com/wuqunyong/file_storage/pkg/encoders"
	"github.com/wuqunyong/file_storage/pkg/errs"
	"github.com/wuqunyong/file_storage/pkg/logger"
	"github.com/wuqunyong/file_storage/rpc_msg"
)

type MsgReq struct {
	TargetId  *concepts.ActorId
	Remote    bool
	SeqId     uint64
	FuncName  uint32
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

func NewMsgReq(target *concepts.ActorId, opcode uint32, args any, ctx context.Context, coder encoders.IEncoder) *MsgReq {
	var cancel context.CancelFunc
	if ctx == nil {
		ctx, cancel = context.WithTimeout(context.Background(), constants.DefaultRPCTimeout)
	} else {
		_, ok := ctx.Deadline()
		if ok {
			ctx, cancel = context.WithCancel(ctx)
		} else {
			ctx, cancel = context.WithTimeout(ctx, constants.DefaultRPCTimeout)
		}
	}

	req := &MsgReq{
		TargetId:  target,
		Remote:    false,
		FuncName:  opcode,
		Args:      args,
		Ctx:       ctx,
		CtxCancel: cancel,
		Done:      make(chan *MsgResp),
		Err:       nil,
		Codec:     coder,
	}
	return req
}

func (req *MsgReq) Error() error {
	return req.Err
}

func (req *MsgReq) GetTarget() *concepts.ActorId {
	return req.TargetId
}

func (req *MsgReq) GetSender() *concepts.ActorId {
	return req.Sender
}

func (req *MsgReq) SetSeqId(value uint64) {
	req.SeqId = value
}

func (req *MsgReq) SetRemote(value bool) error {
	if value {
		req.Remote = value
		data, err := req.Codec.Encode(req.Args)
		if err != nil {
			req.Err = fmt.Errorf("SetRemote failed, err:%s, %w", err.Error(), constants.ErrRPCArgsEncodeFailure)
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

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	select {
	case <-ctx.Done():
		err := ctx.Err()
		logger.Log(logger.ErrorLevel, "MsgReq Send", "data", response, "err", err)
	case req.Done <- response:
	}
}

func (req *MsgReq) HandleResponse(resp concepts.IMsgResp) {
	response, ok := resp.(*MsgResp)
	if !ok {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	select {
	case <-ctx.Done():
		err := ctx.Err()
		logger.Log(logger.ErrorLevel, "MsgReq HandleResponse", "data", response, "err", err)
	case req.Done <- response:
	}
}

func (req *MsgReq) Marshal() ([]byte, error) {
	realm, kind, id, err := concepts.DecodeAddress(req.Sender.Address)
	if err != nil {
		return nil, err
	}

	encoder := encoders.NewProtobufEncoder()
	request := &rpc_msg.RPC_REQUEST{}
	request.Client = &rpc_msg.CLIENT_IDENTIFIER{
		Stub: &rpc_msg.CHANNEL{
			Realm:   realm,
			Type:    kind,
			Id:      id,
			ActorId: req.Sender.ID,
		},
		SeqId:         req.SeqId,
		RequiredReply: true,
		ReplyTopic:    req.Sender.Address,
	}
	request.ServerStream = false
	request.Opcodes = req.FuncName
	request.ArgsData = req.ArgsData

	realm, kind, id, err = concepts.DecodeAddress(req.TargetId.Address)
	if err != nil {
		return nil, err
	}
	request.Server = &rpc_msg.SERVER_IDENTIFIER{
		Stub: &rpc_msg.CHANNEL{
			Realm:   realm,
			Type:    kind,
			Id:      id,
			ActorId: req.TargetId.ID,
		},
	}

	natsRequest := &nats_msg.NATS_MSG_PRXOY{}
	natsRequest.Msg = &nats_msg.NATS_MSG_PRXOY_RpcRequest{
		RpcRequest: request,
	}
	return encoder.Encode(natsRequest)
}

func RequestUnmarshal(data []byte) (*MsgReq, error) {
	encoder := encoders.NewProtobufEncoder()

	natsRequest := &nats_msg.NATS_MSG_PRXOY{}
	err := encoder.Decode(data, natsRequest)
	if err != nil {
		return nil, err
	}

	if natsRequest.GetMsg() == nil {
		return nil, constants.ErrInvalidNatsMsgType
	}

	rpcRequest := natsRequest.GetRpcRequest()
	if rpcRequest == nil {
		return nil, constants.ErrInvalidNatsMsgType
	}
	serverAddress := concepts.GenServerAddress(rpcRequest.Server.Stub.Realm, rpcRequest.Server.Stub.Type, rpcRequest.Server.Stub.Id)
	clientAddress := rpcRequest.Client.ReplyTopic
	if len(clientAddress) == 0 {
		clientAddress = concepts.GenClientAddress(rpcRequest.Client.Stub.Realm, rpcRequest.Client.Stub.Type, rpcRequest.Client.Stub.Id)
	}

	request := NewMsgReq(concepts.NewActorId(serverAddress, rpcRequest.Server.Stub.ActorId), rpcRequest.Opcodes, nil, nil, encoders.NewProtobufEncoder())
	request.Remote = true
	request.SeqId = rpcRequest.GetClient().SeqId
	request.ArgsData = rpcRequest.ArgsData
	request.Sender = concepts.NewActorId(clientAddress, rpcRequest.Client.Stub.ActorId)
	return request, nil
}

func ResponseUnmarshal(data []byte) (*MsgResp, error) {
	natsResponse := &nats_msg.NATS_MSG_PRXOY{}

	encoder := encoders.NewProtobufEncoder()
	err := encoder.Decode(data, natsResponse)
	if err != nil {
		return nil, err
	}

	if natsResponse.GetMsg() == nil {
		return nil, constants.ErrInvalidNatsMsgType
	}

	response := natsResponse.GetRpcResponse()
	if response == nil {
		return nil, constants.ErrInvalidNatsMsgType
	}

	resp := NewMsgResp(response.Client.SeqId, response.Status.Code, response.Status.Msg, encoder)
	resp.ReplyData = response.ResultData
	return resp, nil
}

type MsgResp struct {
	Remote    bool
	SeqId     uint64
	ErrCode   uint32
	ErrMsg    string
	Reply     any
	ReplyData []byte

	Codec encoders.IEncoder
}

func NewMsgResp(seqId uint64, errCode uint32, errMsg string, codec encoders.IEncoder) *MsgResp {
	return &MsgResp{
		SeqId:   seqId,
		ErrCode: errCode,
		ErrMsg:  errMsg,
		Codec:   codec,
	}
}

func (resp *MsgResp) Marshal() ([]byte, error) {
	encoder := encoders.NewProtobufEncoder()

	response := &rpc_msg.RPC_RESPONSE{}
	response.Client = &rpc_msg.CLIENT_IDENTIFIER{
		SeqId: resp.SeqId,
	}
	response.Status = &rpc_msg.STATUS{
		Code: resp.ErrCode,
		Msg:  resp.ErrMsg,
	}
	response.ResultData = resp.ReplyData

	natsResponse := &nats_msg.NATS_MSG_PRXOY{}
	natsResponse.Msg = &nats_msg.NATS_MSG_PRXOY_RpcResponse{
		RpcResponse: response,
	}
	return encoder.Encode(natsResponse)
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
		return nil, errs.NewCodeError(errors.New(response.ErrMsg), int32(response.ErrCode))
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
