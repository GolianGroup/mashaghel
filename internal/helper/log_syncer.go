package helper

import (
	"time"

	"github.com/fluent/fluent-logger-golang/fluent"
)

// FluentBit Write Syncer
type FluentBitWriteSyncer struct {
	logger *fluent.Fluent
	tag    string
}

func NewFluentBitWriteSyncer(host string, port int, tag string) (*FluentBitWriteSyncer, error) {
	fluentLogger, err := fluent.New(
		fluent.Config{
			FluentHost: host,
			FluentPort: port,
		},
	)
	if err != nil {
		return nil, err
	}
	return &FluentBitWriteSyncer{
		logger: fluentLogger,
		tag:    tag,
	}, nil
}

func (f *FluentBitWriteSyncer) Write(p []byte) (n int, err error) {
	logData := map[string]interface{}{
		"timestamp": time.Now().Format(time.RFC3339),
		"message":   string(p),
	}
	err = f.logger.Post(f.tag, logData)
	if err != nil {
		return 0, err
	}
	return len(p), nil
}
