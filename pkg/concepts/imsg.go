package concepts

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/wuqunyong/file_storage/pkg/constants"
)

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

func GenServerAddress(realm, kind, id uint32) string {
	sAddress := fmt.Sprintf("engine.%d.%d.%d.server", realm, kind, id)
	return sAddress
}

func GenClientAddress(realm, kind, id uint32) string {
	sAddress := fmt.Sprintf("engine.%d.%d.%d.client", realm, kind, id)
	return sAddress
}

func DecodeAddress(address string) (realm, kind, id uint32, err error) {
	r := strings.Split(address, ".")
	for _, s := range r {
		if strings.TrimSpace(s) == "" {
			err = constants.ErrInvalidAddress
			return
		}
	}

	if len(r) != 5 {
		err = constants.ErrInvalidAddress
		return
	}

	value, err := strconv.ParseUint(r[1], 10, 32)
	if err != nil {
		return
	}
	realm = uint32(value)

	value, err = strconv.ParseUint(r[2], 10, 32)
	if err != nil {
		return
	}
	kind = uint32(value)

	value, err = strconv.ParseUint(r[3], 10, 32)
	if err != nil {
		return
	}
	id = uint32(value)

	return
}

type IMsgReq interface {
	Marshal() ([]byte, error)
	GetSender() *ActorId
	GetTarget() *ActorId
	SetRemote(value bool) error
	Send(resp IMsgResp)
	SetSeqId(value uint64)
	HandleResponse(resp IMsgResp)
}

type IMsgResp interface {
	Marshal() ([]byte, error)
}
