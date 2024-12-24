package concepts

import "github.com/wuqunyong/file_storage/pkg/common"

type IClientHandler interface {
	Init() error
	CallFunc(request *common.Req)
}
