# gocache-redis v4

[![Go Reference](https://pkg.go.dev/badge/github.com/morkid/gocache-redis/v4.svg)](https://pkg.go.dev/github.com/morkid/gocache-redis/v4)

Redis v4 cache adapter implementing [gocache](https://github.com/morkid/gocache).

## Installation

```bash
go get github.com/morkid/gocache-redis/v4
```

## Example usage

```go
package main

import (
    "fmt"
    "time"

    cache "github.com/morkid/gocache-redis/v4"
    redis "gopkg.in/redis.v4"
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

## License

Published under the [MIT License](../LICENSE).
