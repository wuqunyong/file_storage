package cluster

import (
	"fmt"

	"github.com/nats-io/nats.go"
)

func SetupNatsConn(connectString string, appDieChan chan bool, options ...nats.Option) (*nats.Conn, error) {
	natsOptions := append(
		options,
		nats.DisconnectHandler(func(_ *nats.Conn) {
			fmt.Printf("disconnected from nats!")
		}),
		nats.ReconnectHandler(func(nc *nats.Conn) {
			fmt.Printf("reconnected to nats server %s with address %s in cluster %s!", nc.ConnectedServerName(), nc.ConnectedAddr(), nc.ConnectedClusterName())
		}),
		nats.ClosedHandler(func(nc *nats.Conn) {
			err := nc.LastError()
			if err == nil {
				fmt.Printf("nats connection closed with no error.")
				return
			}

			fmt.Printf("nats connection closed. reason: %q", nc.LastError())
			if appDieChan != nil {
				appDieChan <- true
			}
		}),
		nats.ErrorHandler(func(nc *nats.Conn, sub *nats.Subscription, err error) {
			if err == nats.ErrSlowConsumer {
				dropped, _ := sub.Dropped()
				fmt.Printf("nats slow consumer on subject %q: dropped %d messages\n",
					sub.Subject, dropped)
			} else {
				fmt.Printf(err.Error())
			}
		}),
	)

	nc, err := nats.Connect(connectString, natsOptions...)
	if err != nil {
		return nil, err
	}
	return nc, nil
}
