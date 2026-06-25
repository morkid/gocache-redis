package cache

import (
	"context"
	"errors"
	"time"

	"github.com/morkid/gocache"
	"github.com/redis/go-redis/v9"
)

// RedisCacheConfig base config for redis cache adapter
type RedisCacheConfig struct {
	Client        *redis.Client
	ClusterClient *redis.ClusterClient
	ExpiresIn     time.Duration
	Context       context.Context
}

// NewRedisCache func
func NewRedisCache(config RedisCacheConfig) gocache.AdapterInterface {
	if config.ExpiresIn <= 0 {
		config.ExpiresIn = 3600 * time.Second
	}

	if config.Context == nil {
		config.Context = context.Background()
	}

	if config.Client == nil && config.ClusterClient == nil {
		// No client configured — methods will return errors at runtime
	}

	return &redisCache{
		Client:        config.Client,
		ClusterClient: config.ClusterClient,
		ExpiresIn:     config.ExpiresIn,
		Context:       config.Context,
	}
}

type redisCache struct {
	Client        *redis.Client
	ClusterClient *redis.ClusterClient
	ExpiresIn     time.Duration
	Context       context.Context
}

func (r *redisCache) client() (*redis.Client, *redis.ClusterClient, error) {
	if nil == r.Client && nil == r.ClusterClient {
		return nil, nil, errors.New("Redis client is not defined")
	}
	return r.Client, r.ClusterClient, nil
}

func (r *redisCache) Set(key string, value string) error {
	cli, clusterCli, err := r.client()
	if err != nil {
		return err
	}

	if nil != clusterCli {
		return clusterCli.Set(r.Context, key, value, r.ExpiresIn).Err()
	}

	return cli.Set(r.Context, key, value, r.ExpiresIn).Err()
}

func (r *redisCache) Get(key string) (string, error) {
	cli, clusterCli, err := r.client()
	if err != nil {
		return "", err
	}

	if nil != clusterCli {
		return clusterCli.Get(r.Context, key).Result()
	}

	val, err := cli.Get(r.Context, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return "", errors.New("Cache key not found")
		}
		return "", err
	}

	return val, nil
}

func (r *redisCache) IsValid(key string) bool {
	cli, clusterCli, err := r.client()
	if err != nil {
		return false
	}

	if nil != clusterCli {
		if value, err := clusterCli.Get(r.Context, key).Result(); nil == err && value != "" {
			return true
		}
		return false
	}

	if value, err := cli.Get(r.Context, key).Result(); nil == err && value != "" {
		return true
	}
	return false
}

func (r *redisCache) Clear(key string) error {
	cli, clusterCli, err := r.client()
	if err != nil {
		return err
	}

	if nil != clusterCli {
		return clusterCli.Del(r.Context, key).Err()
	}

	return cli.Del(r.Context, key).Err()
}

func (r *redisCache) ClearPrefix(keyPrefix string) error {
	cli, clusterCli, err := r.client()
	if err != nil {
		return err
	}

	if nil != clusterCli {
		iter := clusterCli.Scan(r.Context, 0, keyPrefix+"*", 0).Iterator()
		for iter.Next(r.Context) {
			clusterCli.Del(r.Context, iter.Val())
		}
		return iter.Err()
	}

	iter := cli.Scan(r.Context, 0, keyPrefix+"*", 0).Iterator()
	for iter.Next(r.Context) {
		cli.Del(r.Context, iter.Val())
	}
	return iter.Err()
}

func (r *redisCache) ClearAll() error {
	cli, clusterCli, err := r.client()
	if err != nil {
		return err
	}

	if nil != clusterCli {
		return clusterCli.FlushDB(r.Context).Err()
	}

	return cli.FlushDB(r.Context).Err()
}
