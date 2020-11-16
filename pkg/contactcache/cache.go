package contactcache

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/spf13/viper"
)

//Cacher interfaces to a caching server such as redis/memcached etc.
type Cacher interface {
	Set(ctx context.Context, key, value string, ttl time.Duration) error
	Get(ctx context.Context, key string) (string, error)
	Delete(ctx context.Context, key string) error
}

//NewRedisCache provides a redis backed cacher
func NewRedisCache() (Cacher, error) {
	redisCache := &RedisCache{}

	endpoint := viper.GetString("cache.address")
	if endpoint == "" {
		return nil, fmt.Errorf("no cache endpoint provided")
	}
	pass := viper.GetString("cache.password")
	db := viper.GetInt("cache.db")

	rdb := redis.NewClient(&redis.Options{
		Addr:     endpoint,
		Password: pass,
		DB:       db,
	})

	redisCache.rdb = rdb

	return redisCache, nil
}

//RedisCache basic redis backend driver
type RedisCache struct {
	rdb *redis.Client
}

//Set sets a cache key
func (rc *RedisCache) Set(ctx context.Context, key, value string, ttl time.Duration) error {
	return rc.rdb.Set(ctx, key, value, ttl).Err()
}

//Get gets a value from the cache
func (rc *RedisCache) Get(ctx context.Context, key string) (string, error) {
	return rc.rdb.Get(ctx, key).Result()
}

//Delete removes a value by key
func (rc *RedisCache) Delete(ctx context.Context, key string) error {
	if strings.Contains(key, "*") {
		return rc.deletePrefix(ctx, key)
	}
	return rc.rdb.Del(ctx, key).Err()
}

func (rc *RedisCache) deletePrefix(ctx context.Context, pattern string) error {
	keys, err := rc.rdb.Keys(ctx, pattern).Result()
	if err != nil {
		return err
	}

	return rc.rdb.Del(ctx, keys...).Err()
}
