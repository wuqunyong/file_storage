package concepts

type IRegistry interface {
	Remove(actorId *ActorId)
	GetByID(id string) IActor //key: ActorId.ID
	GetRootID() []string
}
