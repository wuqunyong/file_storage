package cluster

import (
	"fmt"

	"github.com/nats-io/nats.go"
	"github.com/wuqunyong/file_storage/pkg/logger"
)

func NodeSubject(sRealm, sType, sId string) string {
	return fmt.Sprintf("identify.%s.%s.%s", sRealm, sType, sId)
}

type NatsSubject struct {
	Ch           chan *nats.Msg
	Subject      string
	Subscription *nats.Subscription
}

func NewNatsSubject(subject string, size int) *NatsSubject {
	return &NatsSubject{
		Ch:           make(chan *nats.Msg, size),
		Subject:      subject,
		Subscription: nil,
	}
}

func (p *NatsSubject) Stop() {
	if p.Subscription != nil {
		err := p.Subscription.Unsubscribe()
		if err != nil {
			logger.Log(logger.ErrorLevel, "Unsubscribe error", "subject", p.Subject, "err", err)
		}
	}
	close(p.Ch)
}
