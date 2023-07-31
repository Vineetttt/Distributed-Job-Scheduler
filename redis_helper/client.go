package redishelper

import (
	"github.com/go-redis/redis"
	"github.com/spf13/viper"
)

type RedisCache struct {
	cl *redis.Client
}

func (r RedisCache) HSet(key string, field string, value interface{}) error {
	return r.cl.HSet(key, field, value).Err()
}

func (r RedisCache) HGet(key string, field string) (string, error) {
	return r.cl.HGet(key, field).Result()
}

func (r RedisCache) HExist(key string, field string) (bool, error) {
	return r.cl.HExists(key, field).Result()
}
func CreateNewRedisCache() *RedisCache {
	client := redis.NewClient(&redis.Options{
		Addr:     viper.GetString("REDIS_URL"),
		Password: "",
		DB:       0,
	})

	return &RedisCache{
		cl: client,
	}
}
