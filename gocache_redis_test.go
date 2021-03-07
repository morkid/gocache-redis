package cache_test

import (
	"testing"
	"time"

	cache "github.com/morkid/gocache-redis/v3"
	"gopkg.in/redis.v3"
)

func TestCache(t *testing.T) {
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	config := cache.RedisCacheConfig{
		Client:    client,
		ExpiresIn: 10 * time.Second,
	}

	adapter := *cache.NewRedisCache(config)
	adapter.Set("foo", "bar")

	if adapter.IsValid("foo") {
		value, err := adapter.Get("foo")
		if nil != err {
			t.Error(err)
		} else if value != "bar" {
			t.Error("value not equals to bar")
		} else {
			t.Log(value)
		}
		adapter.Clear("foo")
		if adapter.IsValid("foo") {
			t.Error("Failed to remove key foo")
		}
	}

	if err := adapter.Set("hello", "world"); nil != err {
		t.Error(err)
	}
	if err := adapter.Set("heli", "copter"); nil != err {
		t.Error(err)
	}
	if err := adapter.Set("kitty", "hello"); nil != err {
		t.Error(err)
	}

	if err := adapter.ClearPrefix("hel"); nil != err {
		t.Error(err)
	}

	if value, _ := adapter.Get("heli"); value != "" {
		t.Log(value)
		t.Error("Failed to remove key with prefix hel")
	} else {
		t.Log("All keys with prefix hel was removed")
	}

	if err := adapter.ClearAll(); nil != err {
		t.Error(err)
	}

	if value, _ := adapter.Get("kitty"); value != "" {
		t.Error("Failed to remove all keys")
	}
}
