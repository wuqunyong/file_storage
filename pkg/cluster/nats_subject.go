package cluster

import (
	"fmt"

	"github.com/nats-io/nats.go"
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
	err := p.Subscription.Unsubscribe()
	if err != nil {
		fmt.Printf("Unsubscribe error. [subject = %s, err = %v]", p.Subject, err)
	}
	close(p.Ch)
}
