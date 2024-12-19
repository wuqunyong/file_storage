package api

import (
	"encoding/json"
)

type Encoder interface {
	Encode(data any) ([]byte, error)
	Decode(encodeData []byte, decodeData any) error
}

type JsonEncoder struct {
	// Empty
}

func NewJsonEncoder() *JsonEncoder {
	return &JsonEncoder{}
}

func (je *JsonEncoder) Encode(v any) ([]byte, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (je *JsonEncoder) Decode(data []byte, v any) (err error) {
	return json.Unmarshal(data, v)
}
