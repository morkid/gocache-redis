package cache

import (
	"errors"
	"time"

	"github.com/morkid/gocache"
	"gopkg.in/redis.v3"
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
	}

	return &adapter
}

type redisCache struct {
	Client        *redis.Client
	ClusterClient *redis.ClusterClient
	ExpiresIn     time.Duration
}

func (r redisCache) Set(key string, value string) error {
	if nil == r.Client && nil == r.ClusterClient {
		return r.noClient()
	}
	if err := r.Client.Set(key, value, r.ExpiresIn).Err(); nil != err {
		return err
	}
	return nil
}

func (r redisCache) Get(key string) (string, error) {
	if nil == r.Client && nil == r.ClusterClient {
		return "", r.noClient()
	}
	return r.Client.Get(key).Result()
}

func (r redisCache) IsValid(key string) bool {
	if nil == r.Client && nil == r.ClusterClient {
		return false
	}
	if value, err := r.Client.Get(key).Result(); nil == err && value != "" {
		return true
	}
	return false
}

func (r redisCache) Clear(key string) error {
	if nil == r.Client && nil == r.ClusterClient {
		return r.noClient()
	}

	if nil != r.ClusterClient {
		return r.ClusterClient.Del(key).Err()
	}

	return r.Client.Del(key).Err()
}

func (r redisCache) ClearPrefix(keyPrefix string) error {
	if nil == r.Client && nil == r.ClusterClient {
		return r.noClient()
	}

	var dbSize int64 = r.Client.DbSize().Val()

	if nil != r.ClusterClient {
		_, values, err := r.ClusterClient.Scan(0, keyPrefix+"*", dbSize).Result()
		if nil == err {
			if len(values) < 1 {
				return nil
			}
			return r.ClusterClient.Del(values...).Err()
		}
	}

	_, values, err := r.Client.Scan(0, keyPrefix+"*", dbSize).Result()
	if nil == err {
		if len(values) < 1 {
			return nil
		}
		return r.Client.Del(values...).Err()
	}

	return err
}

func (r redisCache) ClearAll() error {
	if nil == r.Client && nil == r.ClusterClient {
		return r.noClient()
	}

	if nil != r.ClusterClient {
		return r.ClusterClient.FlushDb().Err()
	}

	return r.Client.FlushDb().Err()
}

func (r redisCache) noClient() error {
	return errors.New("Redis client is not defined")
}
