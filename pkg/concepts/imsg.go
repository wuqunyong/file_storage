package concepts

import "github.com/wuqunyong/file_storage/pkg/constants"

type ActorId struct {
	Address string
	ID      string
}

func (actorId *ActorId) String() string {
	return actorId.Address + constants.ActorSeparator + actorId.ID
}

func (actorId *ActorId) GetId() string {
	return actorId.ID
}

func (actorId *ActorId) Equals(other *ActorId) bool {
	return actorId.Address == other.Address && actorId.ID == other.ID
}

func (actorId *ActorId) GenChildId(id string) string {
	childID := actorId.ID + constants.ActorSeparator + id
	return childID
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
