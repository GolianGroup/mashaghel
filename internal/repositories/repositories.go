package repositories

import (
	"context"
	"mashaghel/internal/database/arango"
	"mashaghel/internal/database/scylla"
	"mashaghel/internal/producers"

	"go.uber.org/zap"
)

type Repository interface {
	SystemRepository() SystemRepository
}

// var (
// ErrGlobal = errors.New("some global error")
// )

type repository struct {
	systemRepository SystemRepository
}

func NewRepository(arango arango.ArangoDB, redis producers.RedisClient, scyllaDB scylla.ScyllaDB, logger *zap.Logger, ctx context.Context) Repository {
	systemRepository := NewSystemRepository(arango, redis, scyllaDB)
	return &repository{
		systemRepository: systemRepository,
	}
}

func (r *repository) SystemRepository() SystemRepository {
	return r.systemRepository
}
