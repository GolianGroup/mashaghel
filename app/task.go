package app

import (
	"mashaghel/internal/database/scylla"
	"mashaghel/internal/tasks"

	"go.uber.org/zap"
)

func (a *application) InitTask(scyllaDB scylla.ScyllaDB, logger *zap.Logger) tasks.Task {
	return tasks.NewTaskManager(scyllaDB, logger, &a.config.WorkerPool)
}
