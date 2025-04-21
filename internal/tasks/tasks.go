package tasks

import (
	"fmt"
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
	pool, err := ants.NewPool(t.configs.WorkerPoolSize)
	if err != nil {
		return err
	}
	t.workerpool = pool
	return nil
}

func (t *task) Start() {
	fmt.Println("hala inja")
	go t.watchBackgroundJob()
}

func (t *task) watchBackgroundJob() {

	ticker := time.NewTicker(time.Duration(t.configs.TasksConfig.WatchCooldownDuration) * time.Second)
	defer ticker.Stop()

	longPolling := func() {
		defer func() {
			if r := recover(); r != nil {
				t.logger.Error("panic in run", zap.Any("panic", r))
			}
		}()

		select {
		case <-ticker.C:
			t.processWatched()
		case <-t.quit:
			return
		}
	}

	for {
		longPolling()
	}

}

func (t *task) Stop() {
	t.quit <- struct{}{}
	t.workerpool.Release()
}
