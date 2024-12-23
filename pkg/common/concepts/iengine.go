package concepts

type IEngine interface {
	Request(request IMsgReq) error
	GetAddress() string
	SpawnActor(actor IActor) (*ActorId, error)
	HasActor(id *ActorId) bool
	RemoveActor(id *ActorId)

	AddComponent(component IComponent) error
	HasComponent(name string) bool
	GetComponent(name string) IComponent
}
