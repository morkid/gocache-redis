package cache_test

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redis/v8"
	cache "github.com/morkid/gocache-redis/v8"
)

func setupMiniredis(t *testing.T) (*miniredis.Miniredis, *redis.Client) {
	t.Helper()
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start miniredis: %s", err)
	}
	t.Cleanup(mr.Close)

	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})
	return mr, client
}

// setupClusterClient returns a miniredis instance and a *redis.ClusterClient
// pointing at it, with the ClusterSlots function configured so commands
// bypass the CLUSTER SLOTS discovery handshake. This lets us exercise the
// ClusterClient code paths inside the adapter wrapper functions.
func setupClusterClient(t *testing.T) (*miniredis.Miniredis, *redis.ClusterClient) {
	t.Helper()
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start miniredis: %s", err)
	}
	t.Cleanup(mr.Close)

	client := redis.NewClusterClient(&redis.ClusterOptions{
		// Disable the regular cluster discovery via CLUSTER SLOTS.
		ClusterSlots: func(ctx context.Context) ([]redis.ClusterSlot, error) {
			return []redis.ClusterSlot{
				{
					Start: 0,
					End:   16383,
					Nodes: []redis.ClusterNode{
						{Addr: mr.Addr()},
					},
				},
			}, nil
		},
	})
	return mr, client
}

func TestCache_SetAndGet(t *testing.T) {
	_, client := setupMiniredis(t)

	adapter := cache.NewRedisCache(cache.RedisCacheConfig{
		Client:    client,
		ExpiresIn: 10 * time.Second,
	})

	err := adapter.Set("foo", "bar")
	if err != nil {
		t.Fatalf("unexpected error on Set: %s", err)
	}

	value, err := adapter.Get("foo")
	if err != nil {
		t.Fatalf("unexpected error on Get: %s", err)
	}
	if value != "bar" {
		t.Fatalf("expected 'bar', got '%s'", value)
	}
}

func TestCache_GetNonExistent(t *testing.T) {
	_, client := setupMiniredis(t)

	adapter := cache.NewRedisCache(cache.RedisCacheConfig{
		Client:    client,
		ExpiresIn: 10 * time.Second,
	})

	_, err := adapter.Get("nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent key, got nil")
	}
}

func TestCache_IsValid(t *testing.T) {
	_, client := setupMiniredis(t)

	adapter := cache.NewRedisCache(cache.RedisCacheConfig{
		Client:    client,
		ExpiresIn: 10 * time.Second,
	})

	adapter.Set("foo", "bar")

	if !adapter.IsValid("foo") {
		t.Fatal("expected key to be valid")
	}
	if adapter.IsValid("nonexistent") {
		t.Fatal("expected nonexistent key to be invalid")
	}
}

func TestCache_Clear(t *testing.T) {
	_, client := setupMiniredis(t)

	adapter := cache.NewRedisCache(cache.RedisCacheConfig{
		Client:    client,
		ExpiresIn: 10 * time.Second,
	})

	adapter.Set("foo", "bar")
	adapter.Clear("foo")

	if adapter.IsValid("foo") {
		t.Fatal("expected key to be cleared after Clear")
	}
}

func TestCache_ClearPrefix(t *testing.T) {
	_, client := setupMiniredis(t)

	adapter := cache.NewRedisCache(cache.RedisCacheConfig{
		Client:    client,
		ExpiresIn: 10 * time.Second,
	})

	adapter.Set("hello", "world")
	adapter.Set("heli", "copter")
	adapter.Set("kitty", "hello")

	err := adapter.ClearPrefix("hel")
	if err != nil {
		t.Fatalf("unexpected error on ClearPrefix: %s", err)
	}

	if adapter.IsValid("hello") {
		t.Fatal("expected 'hello' to be cleared")
	}
	if adapter.IsValid("heli") {
		t.Fatal("expected 'heli' to be cleared")
	}
	if !adapter.IsValid("kitty") {
		t.Fatal("expected 'kitty' to remain valid")
	}
}

// TestCache_ClearPrefix_NoMatches covers ClearPrefix when no keys match the
// prefix — the Scan iterator yields no values so the for-loop body is never
// entered and the iter.Err() path returns nil.
func TestCache_ClearPrefix_NoMatches(t *testing.T) {
	_, client := setupMiniredis(t)

	adapter := cache.NewRedisCache(cache.RedisCacheConfig{
		Client:    client,
		ExpiresIn: 10 * time.Second,
	})

	adapter.Set("hello", "world")

	if err := adapter.ClearPrefix("nonexistent_prefix_"); err != nil {
		t.Fatalf("unexpected error on ClearPrefix with no matches: %s", err)
	}

	if !adapter.IsValid("hello") {
		t.Fatal("expected 'hello' to remain valid since prefix didn't match")
	}
}

func TestCache_ClearAll(t *testing.T) {
	_, client := setupMiniredis(t)

	adapter := cache.NewRedisCache(cache.RedisCacheConfig{
		Client:    client,
		ExpiresIn: 10 * time.Second,
	})

	adapter.Set("foo", "bar")
	adapter.Set("baz", "qux")

	err := adapter.ClearAll()
	if err != nil {
		t.Fatalf("unexpected error on ClearAll: %s", err)
	}

	if adapter.IsValid("foo") || adapter.IsValid("baz") {
		t.Fatal("expected all keys to be cleared after ClearAll")
	}
}

func TestCache_MultipleKeys(t *testing.T) {
	_, client := setupMiniredis(t)

	adapter := cache.NewRedisCache(cache.RedisCacheConfig{
		Client:    client,
		ExpiresIn: 10 * time.Second,
	})

	adapter.Set("a", "1")
	adapter.Set("b", "2")
	adapter.Set("c", "3")

	valA, _ := adapter.Get("a")
	valB, _ := adapter.Get("b")
	valC, _ := adapter.Get("c")

	if valA != "1" || valB != "2" || valC != "3" {
		t.Fatalf("unexpected values: a=%s b=%s c=%s", valA, valB, valC)
	}
}

func TestCache_NoClient(t *testing.T) {
	adapter := cache.NewRedisCache(cache.RedisCacheConfig{
		ExpiresIn: 10 * time.Second,
	})

	err := adapter.Set("foo", "bar")
	if err == nil {
		t.Fatal("expected error when no client is set")
	}

	_, err = adapter.Get("foo")
	if err == nil {
		t.Fatal("expected error when no client is set")
	}

	if adapter.IsValid("foo") {
		t.Fatal("expected IsValid to return false when no client is set")
	}

	err = adapter.Clear("foo")
	if err == nil {
		t.Fatal("expected error when no client is set")
	}

	err = adapter.ClearPrefix("foo")
	if err == nil {
		t.Fatal("expected error when no client is set")
	}

	err = adapter.ClearAll()
	if err == nil {
		t.Fatal("expected error when no client is set")
	}
}

// TestCache_DefaultExpiresIn covers NewRedisCache's default ExpiresIn branch
// when the caller passes zero.
func TestCache_DefaultExpiresIn(t *testing.T) {
	_, client := setupMiniredis(t)

	adapter := cache.NewRedisCache(cache.RedisCacheConfig{
		Client: client,
		// ExpiresIn left as zero to trigger default branch.
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

// TestCache_DefaultContext verifies a nil Context in config is replaced with
// context.Background() inside NewRedisCache.
func TestCache_DefaultContext(t *testing.T) {
	_, client := setupMiniredis(t)

	adapter := cache.NewRedisCache(cache.RedisCacheConfig{
		Client:    client,
		ExpiresIn: 10 * time.Second,
		// Context left nil to trigger default branch.
	})

	if err := adapter.Set("foo", "bar"); err != nil {
		t.Fatalf("Set with default Context failed: %s", err)
	}
	if val, err := adapter.Get("foo"); err != nil || val != "bar" {
		t.Fatalf("expected foo=bar, got %s err=%v", val, err)
	}
}

// TestCache_ClientErrorsAfterClose covers the Set/Get/Clear error-propagation
// branches when the underlying *redis.Client is closed. Specifically exercises
// the `Err()/Result()` branches that surface non-nil errors to the caller.
func TestCache_ClientErrorsAfterClose(t *testing.T) {
	mr, client := setupMiniredis(t)

	adapter := cache.NewRedisCache(cache.RedisCacheConfig{
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
	if err := adapter.Clear("any"); err == nil {
		t.Fatal("expected Clear to error after client close")
	}
	if err := adapter.ClearPrefix("any"); err == nil {
		t.Fatal("expected ClearPrefix to error after client close")
	}
	if err := adapter.ClearAll(); err == nil {
		t.Fatal("expected ClearAll to error after client close")
	}
	if adapter.IsValid("any") {
		t.Fatal("expected IsValid to return false after client close")
	}
}

// TestCache_ClusterSuccess exercises every ClusterClient branch in the
// adapter: Set, Get, IsValid, Clear, ClearPrefix, ClearAll.
func TestCache_ClusterSuccess(t *testing.T) {
	mr, clusterCli := setupClusterClient(t)

	adapter := cache.NewRedisCache(cache.RedisCacheConfig{
		ClusterClient: clusterCli,
		ExpiresIn:     10 * time.Second,
	})

	if err := adapter.Set("foo", "bar"); err != nil {
		t.Fatalf("ClusterClient Set failed: %s", err)
	}
	if val, err := adapter.Get("foo"); err != nil || val != "bar" {
		t.Fatalf("ClusterClient Get failed: val=%s err=%v", val, err)
	}
	if !adapter.IsValid("foo") {
		t.Fatal("ClusterClient IsValid returned false after Set")
	}

	// IsValid against a non-existent key exercises the value=="" branch.
	if adapter.IsValid("never_set") {
		t.Fatal("expected IsValid to return false for missing cluster key")
	}

	// ClearPrefix happy path: set multiple keys, scan + delete them.
	_ = adapter.Set("hello", "world")
	_ = adapter.Set("heli", "copter")
	_ = adapter.Set("kitty", "hello")

	if err := adapter.ClearPrefix("hel"); err != nil {
		t.Fatalf("ClusterClient ClearPrefix failed: %s", err)
	}
	if adapter.IsValid("hello") {
		t.Fatal("expected 'hello' to be cleared via cluster")
	}
	if adapter.IsValid("heli") {
		t.Fatal("expected 'heli' to be cleared via cluster")
	}
	if !adapter.IsValid("kitty") {
		t.Fatal("expected 'kitty' to remain valid")
	}

	// ClearPrefix with no matches exercises the iter body skip + iter.Err()==nil branch.
	if err := adapter.ClearPrefix("zzz_nomatch_"); err != nil {
		t.Fatalf("ClusterClient ClearPrefix no-match failed: %s", err)
	}

	if err := adapter.Clear("kitty"); err != nil {
		t.Fatalf("ClusterClient Clear failed: %s", err)
	}
	if adapter.IsValid("kitty") {
		t.Fatal("expected 'kitty' to be cleared")
	}

	_ = adapter.Set("a", "1")
	_ = adapter.Set("b", "2")
	if err := adapter.ClearAll(); err != nil {
		t.Fatalf("ClusterClient ClearAll failed: %s", err)
	}
	if adapter.IsValid("a") || adapter.IsValid("b") {
		t.Fatal("expected all keys to be cleared via cluster ClearAll")
	}

	_ = mr // silence unused
}

// TestCache_ClusterErrorsAfterClose covers the ClusterDel/Scan/Flush
// branches when the cluster client itself is closed. We close the cluster
// client (rather than the underlying miniredis) so that subsequent cleanup
// of miniredis via t.Cleanup doesn't race with an already-detached client.
func TestCache_ClusterErrorsAfterClose(t *testing.T) {
	_, clusterCli := setupClusterClient(t)

	adapter := cache.NewRedisCache(cache.RedisCacheConfig{
		ClusterClient: clusterCli,
		ExpiresIn:     10 * time.Second,
	})

	// Close the cluster client to force subsequent operations to return errors.
	if err := clusterCli.Close(); err != nil {
		t.Fatalf("clusterCli.Close failed: %s", err)
	}

	if err := adapter.Set("foo", "bar"); err == nil {
		t.Fatal("expected ClusterClient Set to error after close")
	}
	if _, err := adapter.Get("any"); err == nil {
		t.Fatal("expected ClusterClient Get to error after close")
	}
	if adapter.IsValid("any") {
		t.Fatal("expected IsValid to return false after cluster close")
	}
	if err := adapter.Clear("any"); err == nil {
		t.Fatal("expected ClusterClient Clear to error after close")
	}
	if err := adapter.ClearPrefix("any"); err == nil {
		t.Fatal("expected ClusterClient ClearPrefix to error after close")
	}
	if err := adapter.ClearAll(); err == nil {
		t.Fatal("expected ClusterClient ClearAll to error after close")
	}
}

func TestCache_ExpiredKey(t *testing.T) {
	mr, client := setupMiniredis(t)

	adapter := cache.NewRedisCache(cache.RedisCacheConfig{
		Client:    client,
		ExpiresIn: 1 * time.Hour,
	})

	adapter.Set("foo", "bar")

	// Force key expiration by manipulating miniredis's internal clock
	mr.FastForward(2 * time.Hour)

	if adapter.IsValid("foo") {
		t.Fatal("expected key to be expired")
	}

	_, err := adapter.Get("foo")
	if err == nil {
		t.Fatal("expected error for expired key")
	}
}

