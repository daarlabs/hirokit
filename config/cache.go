package config

import (
	"github.com/go-redis/redis/v8"
	
	"github.com/daarlabs/hirokit/cache/memory"
)

type Cache struct {
	Memory *memory.Client
	Redis  *redis.Client
}
