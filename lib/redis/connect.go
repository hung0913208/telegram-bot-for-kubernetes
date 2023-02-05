package redis

import (
	redis "github.com/go-redis/redis"
)

type Redis interface {
}

type redisImpl struct {
	client *redis.Client
}
