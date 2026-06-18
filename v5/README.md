# gocache-redis v5

[![Go Reference](https://pkg.go.dev/badge/github.com/morkid/gocache-redis/v5.svg)](https://pkg.go.dev/github.com/morkid/gocache-redis/v5)

Redis v5 cache adapter implementing [gocache](https://github.com/morkid/gocache).

## Installation

```bash
go get github.com/morkid/gocache-redis/v5
```

## Example usage

```go
package main

import (
    "fmt"
    "time"

    cache "github.com/morkid/gocache-redis/v5"
    redis "gopkg.in/redis.v5"
)

func main() {
    client := redis.NewClient(&redis.Options{Addr: "localhost:6379"})

    adapter := *cache.NewRedisCache(cache.RedisCacheConfig{
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
cd v5 && go test -v ./...
```

## License

Published under the [MIT License](../LICENSE).
