package cache

import (
	"context"
	"errors"
	"github.com/SongOf/edge-storage-core/storage"
	"time"

	"github.com/go-redis/redis/v8"
)

type RedisCache struct {
	rdb *redis.Client
}

type RedisOption struct {
	Address  string
	Password string
}

func NewRedisCache(option RedisOption) *RedisCache {
	opt := redis.Options{
		Addr:     option.Address,
		Password: option.Password,
	}
	return &RedisCache{rdb: redis.NewClient(&opt)}
}

func (rc *RedisCache) GetClient() *redis.Client {
	return rc.rdb
}

func (rc *RedisCache) Lock(ctx context.Context, lock, value string, timeout int) error {
	expiration := time.Duration(timeout) * time.Second
	if success, err := rc.rdb.SetNX(ctx, lock, value, expiration).Result(); err != nil {
		storage.CacheErrorInc()
		return err
	} else if !success {
		return errors.New("lock failed")
	}
	return nil
}

func (rc *RedisCache) Unlock(ctx context.Context, lock, value string) error {
	DelScript := redis.NewScript(`
		if redis.call("GET", KEYS[1]) == ARGV[1] then
			return redis.call("DEL", KEYS[1])
		end
		return false
	`)

	if _, err := DelScript.Run(ctx, rc.rdb, []string{lock}, value).Result(); err != nil {
		storage.CacheErrorInc()
		return err
	}
	return nil
}

type RedisClusterOption struct {
	Addresses []string
	User      string
	Password  string
}

type RedisClusterCache struct {
	rdb *redis.ClusterClient
}

func NewRedisClusterCache(option RedisClusterOption) *RedisClusterCache {
	opt := redis.ClusterOptions{
		Addrs: option.Addresses,
	}
	return &RedisClusterCache{
		rdb: redis.NewClusterClient(&opt),
	}
}

func (rcc *RedisClusterCache) GetClient() *redis.ClusterClient {
	return rcc.rdb
}

func (rcc *RedisClusterCache) Lock(ctx context.Context, lock, value string, timeout int) error {
	expiration := time.Duration(timeout) * time.Second
	if err := rcc.rdb.SetNX(ctx, lock, value, expiration).Err(); err != nil {
		storage.CacheErrorInc()
		return err
	}
	return nil
}

func (rcc *RedisClusterCache) Unlock(ctx context.Context, lock, value string) error {
	DelScript := redis.NewScript(`
		if redis.call("GET", KEYS[1]) == ARGV[1] then
			return redis.call("DEL", KEYS[1])
		end
		return false
	`)

	if _, err := DelScript.Run(ctx, rcc.rdb, []string{lock}, value).Result(); err != nil {
		storage.CacheErrorInc()
		return err
	}
	return nil
}
