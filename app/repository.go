package app

import (
	"mashaghel/internal/database/arango"
	"mashaghel/internal/database/scylla"
	"mashaghel/internal/producers"
	"mashaghel/internal/repositories"

	"go.uber.org/zap"
)

func (a *application) InitRepositories(arango arango.ArangoDB, scyllaDB scylla.ScyllaDB, redis producers.RedisClient, logger *zap.Logger) repositories.Repository {
	return repositories.NewRepository(arango, redis, scyllaDB, logger, a.ctx)
}
