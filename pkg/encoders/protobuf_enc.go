package encoders

import (
	"errors"

	"google.golang.org/protobuf/proto"
)

var (
	ErrInvalidProtoMsgEncode = errors.New("Invalid protobuf proto.Message object passed to encode")
	ErrInvalidProtoMsgDecode = errors.New("Invalid protobuf proto.Message object passed to decode")
)

type ProtobufEncoder struct {
	// Empty
}

func NewProtobufEncoder() *ProtobufEncoder {
	return &ProtobufEncoder{}
}

func (pb *ProtobufEncoder) Encode(v any) ([]byte, error) {
	if v == nil {
		return nil, nil
	}
	i, found := v.(proto.Message)
	if !found {
		return nil, ErrInvalidProtoMsgEncode
	}

	b, err := proto.Marshal(i)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (pb *ProtobufEncoder) Decode(data []byte, v any) error {
	i, found := v.(proto.Message)
	if !found {
		return ErrInvalidProtoMsgDecode
	}

	return proto.Unmarshal(data, i)
}
