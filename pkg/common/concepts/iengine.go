package concepts

type IEngine interface {
	Request(request IMsgReq) error
	GetAddress() string
	SpawnActor(actor IActor) (*ActorId, error)
	RemoveActor(id *ActorId)
}
