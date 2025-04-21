package app

import "mashaghel/internal/producers"

func (a *application) InitRedis() producers.RedisClient {
	return producers.NewRedis(&a.config.Redis)
}
