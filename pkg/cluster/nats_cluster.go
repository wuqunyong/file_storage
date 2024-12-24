package cluster

import (
	"log/slog"

	"github.com/nats-io/nats.go"
)

func SetupNatsConn(connectString string, appDieChan chan bool, options ...nats.Option) (*nats.Conn, error) {
	natsOptions := append(
		options,
		nats.DisconnectHandler(func(_ *nats.Conn) {
			slog.Warn("disconnected from nats!")
		}),
		nats.ReconnectHandler(func(nc *nats.Conn) {
			slog.Warn("reconnected to nats", "server", nc.ConnectedServerName(), "address", nc.ConnectedAddr(), "cluster", nc.ConnectedClusterName())
		}),
		nats.ClosedHandler(func(nc *nats.Conn) {
			err := nc.LastError()
			if err == nil {
				slog.Warn("nats connection closed with no error.")
				return
			}

			slog.Warn("nats connection closed.", "reason", nc.LastError())
			if appDieChan != nil {
				appDieChan <- true
			}
		}),
		nats.ErrorHandler(func(nc *nats.Conn, sub *nats.Subscription, err error) {
			if err == nats.ErrSlowConsumer {
				dropped, _ := sub.Dropped()
				slog.Error("nats slow consumer",
					"subject", sub.Subject, "dropped", dropped)
			} else {
				slog.Error(err.Error())
			}
		}),
	)

	nc, err := nats.Connect(connectString, natsOptions...)
	if err != nil {
		return nil, err
	}
	slog.Warn("nats connection success", "address", connectString)
	return nc, nil
}
