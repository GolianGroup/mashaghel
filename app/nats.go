package app

import (
	"mashaghel/internal/helper/nats"

	"go.uber.org/zap"
)

func (a *application) InitNats(logger *zap.Logger) nats.NatsConnection {
	connection, err := nats.NewNatsConnection(a.config.Nats, logger)
	if err != nil {
		logger.Fatal("Fialed to connect to nats", zap.Error(err))
	}

	return connection
}
