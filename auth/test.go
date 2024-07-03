package auth

import (
	"context"
	"testing"
	
	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
	
	"github.com/daarlabs/hirokit/esquel"
)

func createTestDatabaseConnection(t *testing.T) *esquel.DB {
	db, err := esquel.Connect(
		esquel.WithPostgres(),
		esquel.WithHost("localhost"),
		esquel.WithPort(5432),
		esquel.WithDbname("test"),
		esquel.WithUser("cream"),
		esquel.WithPassword("cream"),
		esquel.WithSslDisable(),
	)
	assert.NoError(t, err)
	assert.NoError(t, db.Ping())
	return db
}

func createTestRedisConnection(t *testing.T) *redis.Client {
	client := redis.NewClient(
		&redis.Options{
			Addr: "localhost:6379",
			DB:   10,
		},
	)
	assert.Nil(t, client.Ping(context.Background()).Err())
	return client
}
