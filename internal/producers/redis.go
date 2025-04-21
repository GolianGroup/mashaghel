package producers

import (
	"mashaghel/internal/config"

	"github.com/gofiber/storage/redis/v3"
)

type RedisClient interface {
	RedisStorage() *redis.Storage
	Close() error
}

type Redis struct {
	store *redis.Storage
}

func NewRedis(cfg *config.RedisConfig) RedisClient {

	// Initialize custom config
	store := redis.New(redis.Config{
		Host:      cfg.Host,
		Port:      cfg.Port,
		Password:  cfg.Password,
		Database:  cfg.DB,
		Reset:     false,
		TLSConfig: nil,
		PoolSize:  cfg.PoolSize,
	})
	return &Redis{
		store: store,
	}
}

func (r *Redis) RedisStorage() *redis.Storage {
	return r.store
}

func (r *Redis) Close() error {
	return r.store.Close()
}
