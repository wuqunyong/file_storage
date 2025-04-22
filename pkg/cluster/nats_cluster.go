package cluster

import (
	"github.com/nats-io/nats.go"
	"github.com/wuqunyong/file_storage/pkg/logger"
)

func SetupNatsConn(id, connectString string, appDieChan chan bool, options ...nats.Option) (*nats.Conn, error) {
	natsOptions := append(
		options,
		nats.DisconnectHandler(func(_ *nats.Conn) {
			logger.Log(logger.WarnLevel, "disconnected from nats!", "id", id)
		}),
		nats.ReconnectHandler(func(nc *nats.Conn) {
			logger.Log(logger.WarnLevel, "reconnected to nats", "server", nc.ConnectedServerName(), "id", id, "address", nc.ConnectedAddr(), "cluster", nc.ConnectedClusterName())
		}),
		nats.ClosedHandler(func(nc *nats.Conn) {
			err := nc.LastError()
			if err == nil {
				logger.Log(logger.WarnLevel, "nats connection closed with no error.", "id", id)
				return
			}

			logger.Log(logger.WarnLevel, "nats connection closed.", "id", id, "reason", nc.LastError())
			if appDieChan != nil {
				appDieChan <- true
			}
		}),
		nats.ErrorHandler(func(nc *nats.Conn, sub *nats.Subscription, err error) {
			if err == nats.ErrSlowConsumer {
				dropped, _ := sub.Dropped()
				logger.Log(logger.ErrorLevel, "nats slow consumer",
					"subject", sub.Subject, "dropped", dropped, "id", id)
			} else {
				logger.Log(logger.ErrorLevel, "nats err", "err", err.Error(), "id", id)
			}
		}),
	)

	nc, err := nats.Connect(connectString, natsOptions...)
	if err != nil {
		return nil, err
	}
	logger.Log(logger.WarnLevel, "nats connection success", "id", id, "address", connectString)
	return nc, nil
}
