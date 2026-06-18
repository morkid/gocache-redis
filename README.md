# gocache-redis

[![Go Reference](https://pkg.go.dev/badge/github.com/morkid/gocache-redis/v8.svg)](https://pkg.go.dev/github.com/morkid/gocache-redis/v8)
[![Go Report Card](https://goreportcard.com/badge/github.com/morkid/gocache-redis/v8)](https://goreportcard.com/report/github.com/morkid/gocache-redis/v8)
[![GitHub release (latest SemVer)](https://img.shields.io/github/v/release/morkid/gocache-redis)](https://github.com/morkid/gocache-redis/releases)

Redis-backed cache adapters for [gocache](https://github.com/morkid/gocache), a generic caching library for Go. Supports Redis clients v3 through v8 with offline testing via [miniredis](https://github.com/alicebob/miniredis).

Implements the `gocache.AdapterInterface` with: `Set`, `Get`, `IsValid`, `Clear`, `ClearPrefix`, and `ClearAll`.

## Quick Start

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
    client := redis.NewClient(&redis.Options{Addr: "localhost:6379"})

    adapter := cache.NewRedisCache(cache.RedisCacheConfig{
        Client:    client,
        Context:   context.Background(),
        ExpiresIn: 10 * time.Second,
    })

    adapter.Set("greeting", "hello world")
    value, _ := adapter.Get("greeting")
    fmt.Println(value) // hello world
}
```

## Available Versions

| Version | Redis Client | Module Path |
|---------|-------------|-------------|
| [v8](v8/) | [go-redis/redis v8](https://github.com/go-redis/redis) | `github.com/morkid/gocache-redis/v8` |
| [v7](v7/) | [go-redis/redis v7](https://github.com/go-redis/redis/tree/v7) | `github.com/morkid/gocache-redis/v7` |
| [v5](v5/) | [gopkg.in/redis.v5](https://gopkg.in/redis.v5) | `github.com/morkid/gocache-redis/v5` |
| [v4](v4/) | [gopkg.in/redis.v4](https://gopkg.in/redis.v4) | `github.com/morkid/gocache-redis/v4` |
| [v3](v3/) | [gopkg.in/redis.v3](https://gopkg.in/redis.v3) | `github.com/morkid/gocache-redis/v3` |

## Installation

Choose the version matching your Redis client and run:

```bash
go get github.com/morkid/gocache-redis/v8   # for Redis client v8
go get github.com/morkid/gocache-redis/v7   # for Redis client v7
go get github.com/morkid/gocache-redis/v5   # for Redis client v5
go get github.com/morkid/gocache-redis/v4   # for Redis client v4
go get github.com/morkid/gocache-redis/v3   # for Redis client v3
```

## Development

This repository uses a subdirectory-per-version structure. All versions live on the `master` branch. A `go.work` at the root enables developing all versions simultaneously — no `replace` directives needed.

### Prerequisites

- Go 1.21+ (for workspace development; individual modules require Go 1.16+)
- [gocache](https://github.com/morkid/gocache) v1.0.3 (imported automatically as a dependency)

## License

Published under the [MIT License](LICENSE).
