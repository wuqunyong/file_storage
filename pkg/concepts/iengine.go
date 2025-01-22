package concepts

type IEngine interface {
	Request(request IMsgReq) error
	GetAddress() string
	GetRegistry() IRegistry
	SpawnActor(actor IActor) (*ActorId, error)
	HasActor(id *ActorId) bool
	RemoveActor(id *ActorId)

	MustAddComponent(component IComponent)
	HasComponent(name string) bool
	GetComponent(name string) IComponent

	WaitForShutdown()
}
