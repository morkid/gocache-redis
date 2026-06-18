# gocache-redis v7

[![Go Reference](https://pkg.go.dev/badge/github.com/morkid/gocache-redis/v7.svg)](https://pkg.go.dev/github.com/morkid/gocache-redis/v7)

Redis v7 cache adapter implementing [gocache](https://github.com/morkid/gocache).

## Installation

```bash
go get github.com/morkid/gocache-redis/v7
```

## Example usage

```go
package main

import (
    "fmt"
    "time"

    cache "github.com/morkid/gocache-redis/v7"
    "github.com/go-redis/redis/v7"
)

func main() {
    client := redis.NewClient(&redis.Options{Addr: "localhost:6379"})

    adapter := cache.NewRedisCache(cache.RedisCacheConfig{
        Client:    client,
        ExpiresIn: 10 * time.Second,
    })

    adapter.Set("greeting", "hello world")
    value, _ := adapter.Get("greeting")
    fmt.Println(value) // hello world
}
```

## Testing

Tests use [miniredis](https://github.com/alicebob/miniredis) — no external Redis server required.

```bash
cd v7 && go test -v ./...
```

## License

Published under the [MIT License](../LICENSE).
