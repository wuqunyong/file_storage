package concepts

type IRegistry interface {
	GetActorId(kind, id string) *ActorId
	Remove(actorId *ActorId)
	GetByID(id string) IActor
	GetRootID() []string
}
