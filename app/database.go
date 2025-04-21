package app

import (
	"mashaghel/internal/database/arango"
	"mashaghel/internal/database/scylla"

	"go.uber.org/zap"
)

func (a *application) InitArangoDB(logger *zap.Logger) arango.ArangoDB {
	db, err := arango.NewArangoDB(a.ctx, &a.config.ArangoDB)
	if err != nil {
		logger.Fatal("Failed to start arango database", zap.Error(err))
	}
	return db
}

func (a *application) InitScyllaDB(logger *zap.Logger) scylla.ScyllaDB {
	db, err := scylla.NewScyllaDB(a.ctx, a.config, logger)
	if err != nil {
		logger.Fatal("Failed to start ScyllaDB", zap.Error(err))
	}
	logger.Info("ScyllaDB initialized successfully")
	return db
}
