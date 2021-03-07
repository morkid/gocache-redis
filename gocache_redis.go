package cache

import (
	"context"
	"errors"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/morkid/gocache"
)

// RedisCacheConfig base config for redis cache adapter
type RedisCacheConfig struct {
	Client        *redis.Client
	ClusterClient *redis.ClusterClient
	ExpiresIn     time.Duration
}

// NewRedisCache func
func NewRedisCache(config RedisCacheConfig) *gocache.AdapterInterface {
	if config.ExpiresIn <= 0 {
		config.ExpiresIn = 3600 * time.Second
	}

	var adapter gocache.AdapterInterface = &redisCache{
		Client:        config.Client,
		ClusterClient: config.ClusterClient,
		ExpiresIn:     config.ExpiresIn,
		Context:       context.Background(),
	}

	return &adapter
}

type redisCache struct {
	Client        *redis.Client
	ClusterClient *redis.ClusterClient
	ExpiresIn     time.Duration
	Context       context.Context
}

func (r redisCache) Set(key string, value string) error {
	if nil == r.Client && nil == r.ClusterClient {
		return r.noClient()
	}
	if err := r.Client.Set(r.Context, key, value, r.ExpiresIn).Err(); nil != err {
		return err
	}
	return nil
}

func (r redisCache) Get(key string) (string, error) {
	if nil == r.Client && nil == r.ClusterClient {
		return "", r.noClient()
	}
	return r.Client.Get(r.Context, key).Result()
}

func (r redisCache) IsValid(key string) bool {
	if nil == r.Client && nil == r.ClusterClient {
		return false
	}
	if value, err := r.Client.Get(r.Context, key).Result(); nil == err && value != "" {
		return true
	}
	return false
}

func (r redisCache) Clear(key string) error {
	if nil == r.Client && nil == r.ClusterClient {
		return r.noClient()
	}

	if nil != r.ClusterClient {
		return r.ClusterClient.Del(r.Context, key).Err()
	}

	return r.Client.Del(r.Context, key).Err()
}

func (r redisCache) ClearPrefix(keyPrefix string) error {
	if nil == r.Client && nil == r.ClusterClient {
		return r.noClient()
	}

	if nil != r.ClusterClient {
		values, _, err := r.ClusterClient.Scan(r.Context, 0, keyPrefix+"*", 0).Result()
		if nil == err {
			return r.ClusterClient.Del(r.Context, values...).Err()
		}
	}

	values, _, err := r.Client.Scan(r.Context, 0, keyPrefix+"*", 0).Result()
	if nil == err {
		return r.Client.Del(r.Context, values...).Err()
	}

	return err
}

func (r redisCache) ClearAll() error {
	if nil == r.Client && nil == r.ClusterClient {
		return r.noClient()
	}

	if nil != r.ClusterClient {
		values, _, err := r.ClusterClient.Scan(r.Context, 0, "*", 0).Result()
		if nil == err {
			return r.ClusterClient.Del(r.Context, values...).Err()
		}
	}

	values, _, err := r.Client.Scan(r.Context, 0, "*", 0).Result()
	if nil == err {
		return r.Client.Del(r.Context, values...).Err()
	}

	return err
}

func (r redisCache) noClient() error {
	return errors.New("Redis client is not defined")
}
