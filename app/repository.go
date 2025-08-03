package app

import (
	"mashaghel/internal/database/scylla"
	"mashaghel/internal/repositories"

	"go.uber.org/zap"
)

func (a *application) InitRepositories(scyllaDB scylla.ScyllaDB, logger *zap.Logger) repositories.Repository {
	return repositories.NewRepository(scyllaDB, logger, a.ctx)
}
