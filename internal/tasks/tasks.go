package tasks

import (
	"mashaghel/internal/config"
	"mashaghel/internal/database/scylla"
	"time"

	"github.com/panjf2000/ants/v2"
	"go.uber.org/zap"
)

type Task interface {
	Start()
	Stop()
	InitWorkerPool() error
}

type task struct {
	scylla     scylla.ScyllaDB
	logger     *zap.Logger
	workerpool *ants.Pool
	quit       chan struct{}
	configs    *config.WorkerPoolConfig
}

func NewTaskManager(
	scyllaDB scylla.ScyllaDB,
	logger *zap.Logger,
	configs *config.WorkerPoolConfig,
) Task {
	return &task{
		scylla:     scyllaDB,
		logger:     logger.With(zap.String("task", "watched")),
		workerpool: nil,
		configs:    configs,
		quit:       make(chan struct{}),
	}
}

func (t *task) InitWorkerPool() error {
	t.logger.Info("Initializing worker pool powered by (YA ALI)", zap.Int("size", t.configs.WorkerPoolSize))
	pool, err := ants.NewPool(t.configs.WorkerPoolSize)
	if err != nil {
		t.logger.Error("Failed to initialize worker pool", zap.Error(err))
		return err
	}
	t.workerpool = pool
	t.logger.Info("Worker pool initialized successfully")
	return nil
}

func (t *task) Start() {
	t.logger.Info("Starting task manager")
	go t.watchBackgroundJob()
}

func (t *task) watchBackgroundJob() {
	t.logger.Info("Starting background job watcher")

	ticker := time.NewTicker(time.Duration(t.configs.TasksConfig.WatchCooldownDuration) * time.Second)
	defer ticker.Stop()

	longPolling := func() {
		defer func() {
			if r := recover(); r != nil {
				t.logger.Error("Panic in run", zap.Any("panic", r))
			}
		}()

		select {
		case <-ticker.C:
			t.processWatched()
		case <-t.quit:
			t.logger.Info("Received quit signal, stopping background job watcher")
			return
		}
	}

	for {
		longPolling()
	}
}

func (t *task) Stop() {
	t.logger.Info("Stopping task manager")
	t.quit <- struct{}{}
	t.workerpool.Release()
	t.logger.Info("Task manager stopped successfully")
}
