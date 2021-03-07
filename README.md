# cache redis adapter
[![Go Reference](https://pkg.go.dev/badge/github.com/morkid/gocache-redis/v4.svg)](https://pkg.go.dev/github.com/morkid/gocache-redis/v3)
[![Go](https://github.com/morkid/gocache-redis/actions/workflows/go.yml/badge.svg)](https://github.com/morkid/gocache-redis/actions/workflows/go.yml)
[![Build Status](https://travis-ci.com/morkid/gocache-redis.svg?branch=master)](https://travis-ci.com/morkid/gocache-redis)
[![Go Report Card](https://goreportcard.com/badge/github.com/morkid/gocache-redis/v4)](https://goreportcard.com/report/github.com/morkid/gocache-redis/v4)
[![GitHub release (latest SemVer)](https://img.shields.io/github/v/release/morkid/gocache-redis)](https://github.com/morkid/gocache-redis/releases)

This library is created by implementing [gocache](https://github.com/morkid/gocache) 
and require [redis](https://github.com/go-redis/redis) v4.

## Installation

```bash
go get -d github.com/morkid/gocache-redis/v4
```

Available versions:
- [github.com/morkid/gocache-redis/v8](https://github.com/morkid/gocache-redis/tree/v8) for [redis client v8](https://github.com/go-redis/redis)
- [github.com/morkid/gocache-redis/v7](https://github.com/morkid/gocache-redis/tree/v7) for [redis client v7](https://github.com/go-redis/redis/tree/v7)
- [github.com/morkid/gocache-redis/v5](https://github.com/morkid/gocache-redis/tree/v5) for [redis client v5](https://github.com/go-redis/redis/tree/v5)
- [github.com/morkid/gocache-redis/v4](https://github.com/morkid/gocache-redis/tree/v4) for [redis client v4](https://github.com/go-redis/redis/tree/v4)
- [github.com/morkid/gocache-redis/v3](https://github.com/morkid/gocache-redis/tree/v3) for [redis client v3](https://github.com/morkid/gocache-redis/tree/v3)


## Example usage
```go
package main

import (
    "time"
    "fmt"
    cache "github.com/morkid/gocache-redis/v4"
    "gopkg.in/redis.v4"
)

func main() {
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

## License

Published under the [MIT License](https://github.com/morkid/gocache-redis/blob/master/LICENSE).