package repositories

import (
	"context"
	"mashaghel/internal/database/arango"
	"mashaghel/internal/database/scylla"
	"mashaghel/internal/producers"
)

type SystemRepository interface {
	ArangoPing(ctx context.Context) error
	RedisPing(ctx context.Context) error
}

type systemRepository struct {
	arango   arango.ArangoDB
	redis    producers.RedisClient
	scyllaDB scylla.ScyllaDB
}

func NewSystemRepository(arango arango.ArangoDB, redis producers.RedisClient, scyllaDB scylla.ScyllaDB) SystemRepository {
	return &systemRepository{arango: arango, redis: redis, scyllaDB: scyllaDB}
}

func (r *systemRepository) ArangoPing(ctx context.Context) error {
	if err := r.arango.Ping(ctx); err != nil {
		return err
	}
	return nil
}

func (r *systemRepository) RedisPing(ctx context.Context) error {
	if err := r.redis.RedisStorage().Conn().Ping(ctx).Err(); err != nil {
		return err
	}
	return nil
}

func (r *systemRepository) ScyllaDBPing(ctx context.Context) error {
	if err := r.scyllaDB.Ping(ctx); err != nil {
		return err
	}
	return nil
}
