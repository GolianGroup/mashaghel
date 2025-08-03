package repositories

import (
	"context"
	"mashaghel/internal/database/scylla"

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

func NewRepository(scyllaDB scylla.ScyllaDB, logger *zap.Logger, ctx context.Context) Repository {
	systemRepository := NewSystemRepository(scyllaDB)
	return &repository{
		systemRepository: systemRepository,
	}
}

func (r *repository) SystemRepository() SystemRepository {
	return r.systemRepository
}
