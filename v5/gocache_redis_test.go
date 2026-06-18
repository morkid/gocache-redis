package cache

import (
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	redis "gopkg.in/redis.v5"
)

// NOTE for v5: gopkg.in/redis.v5 ClusterOptions has no ClusterSlots override
// hook. The v5 source Set/Get/IsValid dereference r.Client unconditionally
// (panicking if Client is nil), so cluster tests configure both Client
// and ClusterClient. The Client-only path is fully covered.

func setupMiniredis(t *testing.T) *redis.Client {
	t.Helper()
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start miniredis: %s", err)
	}
	t.Cleanup(mr.Close)

	return redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})
}

func deadClusterClient() *redis.ClusterClient {
	return redis.NewClusterClient(&redis.ClusterOptions{
		Addrs:       []string{"127.0.0.1:1"},
		DialTimeout: 200 * time.Millisecond,
	})
}

func TestCache_NoClient(t *testing.T) {
	adapter := NewRedisCache(RedisCacheConfig{
		ExpiresIn: 10 * time.Second,
	})

	if err := adapter.Set("foo", "bar"); err == nil {
		t.Fatal("expected error when no client is set")
	}
	if _, err := adapter.Get("foo"); err == nil {
		t.Fatal("expected error when no client is set")
	}
	if adapter.IsValid("foo") {
		t.Fatal("expected IsValid to return false when no client is set")
	}
	if err := adapter.Clear("foo"); err == nil {
		t.Fatal("expected error when no client is set")
	}
	if err := adapter.ClearPrefix("foo"); err == nil {
		t.Fatal("expected error when no client is set")
	}
	if err := adapter.ClearAll(); err == nil {
		t.Fatal("expected error when no client is set")
	}
}

func TestCache_DefaultExpiresIn(t *testing.T) {
	client := setupMiniredis(t)

	adapter := NewRedisCache(RedisCacheConfig{
		Client: client,
	})

	if err := adapter.Set("foo", "bar"); err != nil {
		t.Fatalf("Set with default ExpiresIn failed: %s", err)
	}
	if !adapter.IsValid("foo") {
		t.Fatal("expected key to be valid with default ExpiresIn")
	}
	if val, err := adapter.Get("foo"); err != nil || val != "bar" {
		t.Fatalf("expected foo=bar, got %s err=%v", val, err)
	}
}

func TestCache_ClientSuccess(t *testing.T) {
	client := setupMiniredis(t)

	adapter := NewRedisCache(RedisCacheConfig{
		Client:    client,
		ExpiresIn: 10 * time.Second,
	})

	if err := adapter.Set("foo", "bar"); err != nil {
		t.Fatalf("Set failed: %s", err)
	}
	if val, err := adapter.Get("foo"); err != nil || val != "bar" {
		t.Fatalf("Get failed: val=%s err=%v", val, err)
	}
	if !adapter.IsValid("foo") {
		t.Fatal("IsValid returned false after Set")
	}
	if adapter.IsValid("never_set") {
		t.Fatal("expected IsValid false for missing key")
	}

	_ = adapter.Set("hello", "world")
	_ = adapter.Set("heli", "copter")
	_ = adapter.Set("kitty", "hello")

	if err := adapter.ClearPrefix("hel"); err != nil {
		t.Fatalf("ClearPrefix failed: %s", err)
	}
	if adapter.IsValid("hello") || adapter.IsValid("heli") {
		t.Fatal("expected 'hello' and 'heli' to be cleared")
	}
	if !adapter.IsValid("kitty") {
		t.Fatal("expected 'kitty' to remain valid")
	}

	if err := adapter.Clear("kitty"); err != nil {
		t.Fatalf("Clear failed: %s", err)
	}
	if err := adapter.ClearAll(); err != nil {
		t.Fatalf("ClearAll failed: %s", err)
	}
}

func TestCache_ClientClearPrefix_NoMatches(t *testing.T) {
	client := setupMiniredis(t)

	adapter := NewRedisCache(RedisCacheConfig{
		Client:    client,
		ExpiresIn: 10 * time.Second,
	})

	_ = adapter.Set("hello", "world")

	if err := adapter.ClearPrefix("zzz_nomatch_"); err != nil {
		t.Fatalf("ClearPrefix no-match failed: %s", err)
	}
	if !adapter.IsValid("hello") {
		t.Fatal("'hello' should remain valid since prefix didn't match")
	}
}

func TestCache_ClientErrorsAfterClose(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start miniredis: %s", err)
	}
	defer mr.Close()

	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})

	adapter := NewRedisCache(RedisCacheConfig{
		Client:    client,
		ExpiresIn: 10 * time.Second,
	})

	mr.Close()

	if err := adapter.Set("foo", "bar"); err == nil {
		t.Fatal("expected Set to error after client close")
	}
	if _, err := adapter.Get("any"); err == nil {
		t.Fatal("expected Get to error after client close")
	}
	if adapter.IsValid("any") {
		t.Fatal("expected IsValid to return false after client close")
	}
	if err := adapter.Clear("any"); err == nil {
		t.Fatal("expected Clear to error after client close")
	}
	if err := adapter.ClearPrefix("any"); err == nil {
		t.Fatal("expected ClearPrefix to error after client close")
	}
	if err := adapter.ClearAll(); err == nil {
		t.Fatal("expected ClearAll to error after client close")
	}
}

func TestCache_ClusterErrorsAfterClose(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start miniredis: %s", err)
	}
	defer mr.Close()

	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	clusterCli := deadClusterClient()

	adapter := NewRedisCache(RedisCacheConfig{
		Client:        client,
		ClusterClient: clusterCli,
		ExpiresIn:     10 * time.Second,
	})

	mr.Close()

	if err := adapter.Set("foo", "bar"); err == nil {
		t.Fatal("expected Set to error after close")
	}
	if _, err := adapter.Get("any"); err == nil {
		t.Fatal("expected Get to error after close")
	}
	if adapter.IsValid("any") {
		t.Fatal("expected IsValid to return false after close")
	}
	if err := adapter.Clear("any"); err == nil {
		t.Fatal("expected Clear to error after close (cluster branch)")
	}
	if err := adapter.ClearPrefix("any"); err == nil {
		t.Fatal("expected ClearPrefix to error after close (cluster branch)")
	}
	if err := adapter.ClearAll(); err == nil {
		t.Fatal("expected ClearAll to error after close (cluster branch)")
	}
}
