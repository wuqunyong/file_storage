package concepts

const actorSeparator = "."

type ActorId struct {
	Address string
	ID      string
}

func (actorId *ActorId) String() string {
	return actorId.Address + actorSeparator + actorId.ID
}

func (actorId *ActorId) GetId() string {
	return actorId.ID
}

func (actorId *ActorId) Equals(other *ActorId) bool {
	return actorId.Address == other.Address && actorId.ID == other.ID
}

func NewActorId(address, id string) *ActorId {
	actorId := &ActorId{
		Address: address,
		ID:      id,
	}
	return actorId
}

type IMsgReq interface {
	Marshal() ([]byte, error)
	GetSender() *ActorId
	GetTarget() *ActorId
	SetRemote(value bool) error
	Send(resp IMsgResp)
	SetSeqId(value int64)
	HandleResponse(resp IMsgResp)
}

type IMsgResp interface {
	Marshal() ([]byte, error)
}
