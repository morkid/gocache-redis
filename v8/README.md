# gocache-redis v8

[![Go Reference](https://pkg.go.dev/badge/github.com/morkid/gocache-redis/v8.svg)](https://pkg.go.dev/github.com/morkid/gocache-redis/v8)

Redis v8 cache adapter implementing [gocache](https://github.com/morkid/gocache).

## Installation

```bash
go get github.com/morkid/gocache-redis/v8
```

## Example usage

```go
package main

import (
    "context"
    "fmt"
    "time"

    cache "github.com/morkid/gocache-redis/v8"
    "github.com/go-redis/redis/v8"
)

func main() {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    client := redis.NewClient(&redis.Options{
        Addr:     "localhost:6379",
        Password: "",
        DB:       0,
    })

    config := cache.RedisCacheConfig{
        Client:    client,
        Context:   ctx,
        ExpiresIn: 10 * time.Second,
    }

    adapter := cache.NewRedisCache(config)
    adapter.Set("foo", "bar")

    if adapter.IsValid("foo") {
        value, err := adapter.Get("foo")
        if err != nil {
            fmt.Println(err)
        } else if value != "bar" {
            fmt.Println("value not equals to bar")
        } else {
            fmt.Println(value)
        }
        adapter.Clear("foo")
        if adapter.IsValid("foo") {
            fmt.Println("Failed to remove key foo")
        }
    }
}
```

## Testing

Tests use [miniredis](https://github.com/alicebob/miniredis) — no external Redis server required.

```bash
cd v8 && go test -v ./...
```

## License

Published under the [MIT License](../LICENSE).
