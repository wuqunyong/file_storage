// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.33.0
// 	protoc        v4.25.3
// source: proto/nats_msg/nats_msg.proto

package nats_msg

import (
	rpc_msg "github.com/wuqunyong/file_storage/proto/rpc_msg"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type NATS_MSG_PRXOY struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Types that are assignable to Msg:
	//
	//	*NATS_MSG_PRXOY_RpcRequest
	//	*NATS_MSG_PRXOY_RpcResponse
	//	*NATS_MSG_PRXOY_MultiplexerForward
	//	*NATS_MSG_PRXOY_DemultiplexerForward
	Msg isNATS_MSG_PRXOY_Msg `protobuf_oneof:"msg"`
}

func (x *NATS_MSG_PRXOY) Reset() {
	*x = NATS_MSG_PRXOY{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_nats_msg_nats_msg_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *NATS_MSG_PRXOY) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*NATS_MSG_PRXOY) ProtoMessage() {}

func (x *NATS_MSG_PRXOY) ProtoReflect() protoreflect.Message {
	mi := &file_proto_nats_msg_nats_msg_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use NATS_MSG_PRXOY.ProtoReflect.Descriptor instead.
func (*NATS_MSG_PRXOY) Descriptor() ([]byte, []int) {
	return file_proto_nats_msg_nats_msg_proto_rawDescGZIP(), []int{0}
}

func (m *NATS_MSG_PRXOY) GetMsg() isNATS_MSG_PRXOY_Msg {
	if m != nil {
		return m.Msg
	}
	return nil
}

func (x *NATS_MSG_PRXOY) GetRpcRequest() *rpc_msg.RPC_REQUEST {
	if x, ok := x.GetMsg().(*NATS_MSG_PRXOY_RpcRequest); ok {
		return x.RpcRequest
	}
	return nil
}

func (x *NATS_MSG_PRXOY) GetRpcResponse() *rpc_msg.RPC_RESPONSE {
	if x, ok := x.GetMsg().(*NATS_MSG_PRXOY_RpcResponse); ok {
		return x.RpcResponse
	}
	return nil
}

func (x *NATS_MSG_PRXOY) GetMultiplexerForward() *rpc_msg.RPC_Multiplexer_Forward {
	if x, ok := x.GetMsg().(*NATS_MSG_PRXOY_MultiplexerForward); ok {
		return x.MultiplexerForward
	}
	return nil
}

func (x *NATS_MSG_PRXOY) GetDemultiplexerForward() *rpc_msg.PRC_DeMultiplexer_Forward {
	if x, ok := x.GetMsg().(*NATS_MSG_PRXOY_DemultiplexerForward); ok {
		return x.DemultiplexerForward
	}
	return nil
}

type isNATS_MSG_PRXOY_Msg interface {
	isNATS_MSG_PRXOY_Msg()
}

type NATS_MSG_PRXOY_RpcRequest struct {
	RpcRequest *rpc_msg.RPC_REQUEST `protobuf:"bytes,100,opt,name=rpc_request,json=rpcRequest,proto3,oneof"`
}

type NATS_MSG_PRXOY_RpcResponse struct {
	RpcResponse *rpc_msg.RPC_RESPONSE `protobuf:"bytes,101,opt,name=rpc_response,json=rpcResponse,proto3,oneof"`
}

type NATS_MSG_PRXOY_MultiplexerForward struct {
	MultiplexerForward *rpc_msg.RPC_Multiplexer_Forward `protobuf:"bytes,102,opt,name=multiplexer_forward,json=multiplexerForward,proto3,oneof"`
}

type NATS_MSG_PRXOY_DemultiplexerForward struct {
	DemultiplexerForward *rpc_msg.PRC_DeMultiplexer_Forward `protobuf:"bytes,103,opt,name=demultiplexer_forward,json=demultiplexerForward,proto3,oneof"`
}

func (*NATS_MSG_PRXOY_RpcRequest) isNATS_MSG_PRXOY_Msg() {}

func (*NATS_MSG_PRXOY_RpcResponse) isNATS_MSG_PRXOY_Msg() {}

func (*NATS_MSG_PRXOY_MultiplexerForward) isNATS_MSG_PRXOY_Msg() {}

func (*NATS_MSG_PRXOY_DemultiplexerForward) isNATS_MSG_PRXOY_Msg() {}

var File_proto_nats_msg_nats_msg_proto protoreflect.FileDescriptor

var file_proto_nats_msg_nats_msg_proto_rawDesc = []byte{
	0x0a, 0x1d, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x6e, 0x61, 0x74, 0x73, 0x5f, 0x6d, 0x73, 0x67,
	0x2f, 0x6e, 0x61, 0x74, 0x73, 0x5f, 0x6d, 0x73, 0x67, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12,
	0x08, 0x6e, 0x61, 0x74, 0x73, 0x5f, 0x6d, 0x73, 0x67, 0x1a, 0x1b, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x2f, 0x72, 0x70, 0x63, 0x5f, 0x6d, 0x73, 0x67, 0x2f, 0x72, 0x70, 0x63, 0x5f, 0x6d, 0x73, 0x67,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0xbc, 0x02, 0x0a, 0x0e, 0x4e, 0x41, 0x54, 0x53, 0x5f,
	0x4d, 0x53, 0x47, 0x5f, 0x50, 0x52, 0x58, 0x4f, 0x59, 0x12, 0x37, 0x0a, 0x0b, 0x72, 0x70, 0x63,
	0x5f, 0x72, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x18, 0x64, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x14,
	0x2e, 0x72, 0x70, 0x63, 0x5f, 0x6d, 0x73, 0x67, 0x2e, 0x52, 0x50, 0x43, 0x5f, 0x52, 0x45, 0x51,
	0x55, 0x45, 0x53, 0x54, 0x48, 0x00, 0x52, 0x0a, 0x72, 0x70, 0x63, 0x52, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x12, 0x3a, 0x0a, 0x0c, 0x72, 0x70, 0x63, 0x5f, 0x72, 0x65, 0x73, 0x70, 0x6f, 0x6e,
	0x73, 0x65, 0x18, 0x65, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x15, 0x2e, 0x72, 0x70, 0x63, 0x5f, 0x6d,
	0x73, 0x67, 0x2e, 0x52, 0x50, 0x43, 0x5f, 0x52, 0x45, 0x53, 0x50, 0x4f, 0x4e, 0x53, 0x45, 0x48,
	0x00, 0x52, 0x0b, 0x72, 0x70, 0x63, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x53,
	0x0a, 0x13, 0x6d, 0x75, 0x6c, 0x74, 0x69, 0x70, 0x6c, 0x65, 0x78, 0x65, 0x72, 0x5f, 0x66, 0x6f,
	0x72, 0x77, 0x61, 0x72, 0x64, 0x18, 0x66, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x20, 0x2e, 0x72, 0x70,
	0x63, 0x5f, 0x6d, 0x73, 0x67, 0x2e, 0x52, 0x50, 0x43, 0x5f, 0x4d, 0x75, 0x6c, 0x74, 0x69, 0x70,
	0x6c, 0x65, 0x78, 0x65, 0x72, 0x5f, 0x46, 0x6f, 0x72, 0x77, 0x61, 0x72, 0x64, 0x48, 0x00, 0x52,
	0x12, 0x6d, 0x75, 0x6c, 0x74, 0x69, 0x70, 0x6c, 0x65, 0x78, 0x65, 0x72, 0x46, 0x6f, 0x72, 0x77,
	0x61, 0x72, 0x64, 0x12, 0x59, 0x0a, 0x15, 0x64, 0x65, 0x6d, 0x75, 0x6c, 0x74, 0x69, 0x70, 0x6c,
	0x65, 0x78, 0x65, 0x72, 0x5f, 0x66, 0x6f, 0x72, 0x77, 0x61, 0x72, 0x64, 0x18, 0x67, 0x20, 0x01,
	0x28, 0x0b, 0x32, 0x22, 0x2e, 0x72, 0x70, 0x63, 0x5f, 0x6d, 0x73, 0x67, 0x2e, 0x50, 0x52, 0x43,
	0x5f, 0x44, 0x65, 0x4d, 0x75, 0x6c, 0x74, 0x69, 0x70, 0x6c, 0x65, 0x78, 0x65, 0x72, 0x5f, 0x46,
	0x6f, 0x72, 0x77, 0x61, 0x72, 0x64, 0x48, 0x00, 0x52, 0x14, 0x64, 0x65, 0x6d, 0x75, 0x6c, 0x74,
	0x69, 0x70, 0x6c, 0x65, 0x78, 0x65, 0x72, 0x46, 0x6f, 0x72, 0x77, 0x61, 0x72, 0x64, 0x42, 0x05,
	0x0a, 0x03, 0x6d, 0x73, 0x67, 0x42, 0x3b, 0x5a, 0x39, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e,
	0x63, 0x6f, 0x6d, 0x2f, 0x77, 0x75, 0x71, 0x75, 0x6e, 0x79, 0x6f, 0x6e, 0x67, 0x2f, 0x66, 0x69,
	0x6c, 0x65, 0x5f, 0x73, 0x74, 0x6f, 0x72, 0x61, 0x67, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x2f, 0x6e, 0x61, 0x74, 0x73, 0x5f, 0x6d, 0x73, 0x67, 0x3b, 0x6e, 0x61, 0x74, 0x73, 0x5f, 0x6d,
	0x73, 0x67, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_proto_nats_msg_nats_msg_proto_rawDescOnce sync.Once
	file_proto_nats_msg_nats_msg_proto_rawDescData = file_proto_nats_msg_nats_msg_proto_rawDesc
)

func file_proto_nats_msg_nats_msg_proto_rawDescGZIP() []byte {
	file_proto_nats_msg_nats_msg_proto_rawDescOnce.Do(func() {
		file_proto_nats_msg_nats_msg_proto_rawDescData = protoimpl.X.CompressGZIP(file_proto_nats_msg_nats_msg_proto_rawDescData)
	})
	return file_proto_nats_msg_nats_msg_proto_rawDescData
}

var file_proto_nats_msg_nats_msg_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_proto_nats_msg_nats_msg_proto_goTypes = []interface{}{
	(*NATS_MSG_PRXOY)(nil),                    // 0: nats_msg.NATS_MSG_PRXOY
	(*rpc_msg.RPC_REQUEST)(nil),               // 1: rpc_msg.RPC_REQUEST
	(*rpc_msg.RPC_RESPONSE)(nil),              // 2: rpc_msg.RPC_RESPONSE
	(*rpc_msg.RPC_Multiplexer_Forward)(nil),   // 3: rpc_msg.RPC_Multiplexer_Forward
	(*rpc_msg.PRC_DeMultiplexer_Forward)(nil), // 4: rpc_msg.PRC_DeMultiplexer_Forward
}
var file_proto_nats_msg_nats_msg_proto_depIdxs = []int32{
	1, // 0: nats_msg.NATS_MSG_PRXOY.rpc_request:type_name -> rpc_msg.RPC_REQUEST
	2, // 1: nats_msg.NATS_MSG_PRXOY.rpc_response:type_name -> rpc_msg.RPC_RESPONSE
	3, // 2: nats_msg.NATS_MSG_PRXOY.multiplexer_forward:type_name -> rpc_msg.RPC_Multiplexer_Forward
	4, // 3: nats_msg.NATS_MSG_PRXOY.demultiplexer_forward:type_name -> rpc_msg.PRC_DeMultiplexer_Forward
	4, // [4:4] is the sub-list for method output_type
	4, // [4:4] is the sub-list for method input_type
	4, // [4:4] is the sub-list for extension type_name
	4, // [4:4] is the sub-list for extension extendee
	0, // [0:4] is the sub-list for field type_name
}

func init() { file_proto_nats_msg_nats_msg_proto_init() }
func file_proto_nats_msg_nats_msg_proto_init() {
	if File_proto_nats_msg_nats_msg_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_proto_nats_msg_nats_msg_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*NATS_MSG_PRXOY); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	file_proto_nats_msg_nats_msg_proto_msgTypes[0].OneofWrappers = []interface{}{
		(*NATS_MSG_PRXOY_RpcRequest)(nil),
		(*NATS_MSG_PRXOY_RpcResponse)(nil),
		(*NATS_MSG_PRXOY_MultiplexerForward)(nil),
		(*NATS_MSG_PRXOY_DemultiplexerForward)(nil),
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_proto_nats_msg_nats_msg_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_proto_nats_msg_nats_msg_proto_goTypes,
		DependencyIndexes: file_proto_nats_msg_nats_msg_proto_depIdxs,
		MessageInfos:      file_proto_nats_msg_nats_msg_proto_msgTypes,
	}.Build()
	File_proto_nats_msg_nats_msg_proto = out.File
	file_proto_nats_msg_nats_msg_proto_rawDesc = nil
	file_proto_nats_msg_nats_msg_proto_goTypes = nil
	file_proto_nats_msg_nats_msg_proto_depIdxs = nil
}
