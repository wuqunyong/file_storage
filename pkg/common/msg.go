package common

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
