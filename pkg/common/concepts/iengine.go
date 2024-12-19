package concepts

type IEngine interface {
	Request(request IMsgReq) error
}
