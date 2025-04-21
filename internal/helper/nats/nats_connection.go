package nats

import (
	"mashaghel/internal/config"
	"time"

	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
)

type NatsConnection interface {
	Close()
	Publish(subject string, message []byte) error
}

type natsConnection struct {
	nc     *nats.Conn
	js     nats.JetStreamContext
	logger *zap.Logger
}

func NewNatsConnection(conf config.NatsConfig, logger *zap.Logger) (NatsConnection, error) {
	logger.With(zap.String("nats connection", config.GetNatsURL(&conf)))

	nc, err := nats.Connect(
		config.GetNatsURL(&conf),
		nats.UserInfo(conf.Username, conf.Password),
		nats.Timeout(100*time.Second),
	)
	if err != nil {
		logger.Error("Failed to connect to nats",
			zap.Error(err),
			zap.String("nats connection", config.GetNatsURL(&conf)),
		)
		return nil, err
	}
	logger.Info("Connected to nats successfully.")

	js, err := nc.JetStream()
	if err != nil {
		logger.Error("Failed to create jetstream", zap.Error(err))
		return nil, err
	}

	streamName := conf.StreamConfig.Name
	streamInfo, err := js.StreamInfo(streamName)
	if err != nil && err != nats.ErrStreamNotFound {
		logger.Error("Error getting stream info", zap.Error(err))
		return nil, err
	}

	if streamInfo != nil {
		logger.Info("Stream already exists", zap.String("stream", streamName))
	} else {
		logger.Info("Stream does not exist, creating new stream", zap.String("stream", streamName))

		// Stream config
		var retention nats.RetentionPolicy
		switch conf.StreamConfig.Retention {
		case "limits":
			retention = nats.LimitsPolicy
		case "interest":
			retention = nats.InterestPolicy
		case "workqueue":
			retention = nats.WorkQueuePolicy
		}

		var discard nats.DiscardPolicy
		switch conf.StreamConfig.Discard {
		case "old":
			discard = nats.DiscardOld
		case "new":
			discard = nats.DiscardNew
		}

		var storage nats.StorageType
		switch conf.StreamConfig.Storage {
		case "file":
			storage = nats.FileStorage
		case "memory":
			storage = nats.MemoryStorage
		}

		streamCfg := nats.StreamConfig{
			Name:         streamName,
			Subjects:     []string{"likes.change"},
			Retention:    retention,
			Discard:      discard,
			Storage:      storage,
			NoAck:        true,
			MaxConsumers: conf.StreamConfig.MaxConsumers,
			MaxAge:       conf.StreamConfig.MaxAge * time.Minute,
		}

		_, err := js.AddStream(&streamCfg)
		if err != nil {
			logger.Error("Failed to add stream", zap.Error(err))
			return nil, err
		}

		logger.Info("Stream created successfully", zap.String("stream", streamName))
	}

	return &natsConnection{
		nc:     nc,
		js:     js,
		logger: logger,
	}, nil
}

func (n *natsConnection) Close() {
	n.nc.Close()
}

func (n *natsConnection) Publish(subject string, message []byte) error {
	n.logger.Info("Publishing message ... ",
		zap.String("subject", subject),
		zap.String("message", string(message)),
	)

	_, err := n.js.PublishAsync(subject, message)
	if err != nil {
		n.logger.Error("Failed to publish message",
			zap.Error(err),
			zap.String("subject", subject),
			zap.String("message", string(message)),
		)
		return err
	}

	return nil
}
