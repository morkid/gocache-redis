package cache_test

import (
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redis/v7"
	cache "github.com/morkid/gocache-redis/v7"
)

// NOTE for v7:
//  1. Set/Get/IsValid in the v7 source dereference r.Client unconditionally
//     even when ClusterClient is set, so cluster-only config would panic.
//     Tests therefore combine Client + ClusterClient.
//  2. TestCache_ClusterSuccess uses miniredis with a ClusterSlots override
//     (no context parameter on v7) to drive the cluster happy path. The
//     override captures the addr string at setup time so subsequent mr
//     operations never dereference mr.
//  3. TestCache_ClusterErrorsAfterClose uses a DeadClusterClient that never
//     touches miniredis at all. The cluster branch in Clear/ClearPrefix/
//     ClearAll still fires (ClusterClient != nil) and propagates dial errors.

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

// setupClusterWithMiniredis starts a miniredis, configures a ClusterClient
// with a ClusterSlots handler pointing at the captured addr string, and
// returns both handles. The addr is captured as a string so that operations
// after mr.Close() still see a valid (but unreachable) destination instead
// of dereferencing mr internals and panicking.
func setupClusterWithMiniredis(t *testing.T) (*miniredis.Miniredis, *redis.ClusterClient) {
	t.Helper()
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start miniredis: %s", err)
	}
	t.Cleanup(mr.Close)

	addr := mr.Addr() // capture once before any close
	clusterCli := redis.NewClusterClient(&redis.ClusterOptions{
		ClusterSlots: func() ([]redis.ClusterSlot, error) {
			return []redis.ClusterSlot{
				{
					Start: 0,
					End:   16383,
					Nodes: []redis.ClusterNode{
						{Addr: addr},
					},
				},
			}, nil
		},
	})
	return mr, clusterCli
}

// deadClusterClient returns a ClusterClient whose Addrs point at an
// unreachable address with a short DialTimeout. The cluster remains alive
// (lazy connection) but every operation fails fast with a dial error.
func deadClusterClient() *redis.ClusterClient {
	return redis.NewClusterClient(&redis.ClusterOptions{
		Addrs:       []string{"127.0.0.1:1"},
		DialTimeout: 200 * time.Millisecond,
	})
}

func TestCache_NoClient(t *testing.T) {
	adapter := cache.NewRedisCache(cache.RedisCacheConfig{
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

	adapter := cache.NewRedisCache(cache.RedisCacheConfig{
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

	adapter := cache.NewRedisCache(cache.RedisCacheConfig{
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

	adapter := cache.NewRedisCache(cache.RedisCacheConfig{
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

// TestCache_ClusterSuccess exercises the cluster happy path. mr stays alive
// throughout the test; only the deferred cleanup closes it.
func TestCache_ClusterSuccess(t *testing.T) {
	mr, clusterCli := setupClusterWithMiniredis(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})

	adapter := cache.NewRedisCache(cache.RedisCacheConfig{
		Client:        client,
		ClusterClient: clusterCli,
		ExpiresIn:     10 * time.Second,
	})

	// Set/Get/IsValid via Client.
	if err := adapter.Set("foo", "bar"); err != nil {
		t.Fatalf("Set failed: %s", err)
	}
	if val, err := adapter.Get("foo"); err != nil || val != "bar" {
		t.Fatalf("Get failed: val=%s err=%v", val, err)
	}
	if !adapter.IsValid("foo") {
		t.Fatal("IsValid returned false after Set")
	}

	// Populate additional keys for ClearPrefix (cluster path).
	if err := adapter.Set("hello", "world"); err != nil {
		t.Fatalf("Set 'hello' failed: %s", err)
	}
	if err := adapter.Set("heli", "copter"); err != nil {
		t.Fatalf("Set 'heli' failed: %s", err)
	}
	if err := adapter.Set("kitty", "hello"); err != nil {
		t.Fatalf("Set 'kitty' failed: %s", err)
	}

	// ClearPrefix via cluster: scan yields "hello" and "heli", Del called.
	if err := adapter.ClearPrefix("hel"); err != nil {
		t.Fatalf("Cluster ClearPrefix failed: %s", err)
	}
	if adapter.IsValid("hello") || adapter.IsValid("heli") {
		t.Fatal("expected 'hello' and 'heli' to be cleared via cluster")
	}
	if !adapter.IsValid("kitty") {
		t.Fatal("expected 'kitty' to remain valid")
	}

	if err := adapter.ClearPrefix("zzz_nomatch_"); err != nil {
		t.Fatalf("Cluster ClearPrefix no-match failed: %s", err)
	}

	// Clear and ClearAll via cluster.
	if err := adapter.Clear("kitty"); err != nil {
		t.Fatalf("Cluster Clear failed: %s", err)
	}
	if err := adapter.Set("a", "1"); err != nil {
		t.Fatalf("Set 'a' failed: %s", err)
	}
	if err := adapter.Set("b", "2"); err != nil {
		t.Fatalf("Set 'b' failed: %s", err)
	}
	if err := adapter.ClearAll(); err != nil {
		t.Fatalf("Cluster ClearAll failed: %s", err)
	}
}

// TestCache_ClusterErrorsAfterClose uses a DeadClusterClient (no miniredis
// involved) so the cluster branch in Clear/ClearPrefix/ClearAll still
// fires but every operation propagates a dial error.
func TestCache_ClusterErrorsAfterClose(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start miniredis: %s", err)
	}
	defer mr.Close()

	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	clusterCli := deadClusterClient()

	adapter := cache.NewRedisCache(cache.RedisCacheConfig{
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
