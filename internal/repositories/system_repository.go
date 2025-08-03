package repositories

import (
	"context"
	"mashaghel/internal/database/scylla"
)

type SystemRepository interface {
	// ArangoPing(ctx context.Context) error
	ScyllaDBPing(ctx context.Context) error
	// RedisPing(ctx context.Context) error
}

type systemRepository struct {
	// arango   arango.ArangoDB
	// redis    producers.RedisClient
	scyllaDB scylla.ScyllaDB
}

func NewSystemRepository(scyllaDB scylla.ScyllaDB) SystemRepository {
	return &systemRepository{scyllaDB: scyllaDB}
}

// func (r *systemRepository) ArangoPing(ctx context.Context) error {
// 	if err := r.arango.Ping(ctx); err != nil {
// 		return err
// 	}
// 	return nil
// }

// func (r *systemRepository) RedisPing(ctx context.Context) error {
// 	if err := r.redis.RedisStorage().Conn().Ping(ctx).Err(); err != nil {
// 		return err
// 	}
// 	return nil
// }

func (r *systemRepository) ScyllaDBPing(ctx context.Context) error {
	if err := r.scyllaDB.Ping(ctx); err != nil {
		return err
	}
	return nil
}
