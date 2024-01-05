package kvstorages

import (
	"context"
	"time"

	"github.com/ocontest/backend/pkg"
	"github.com/redis/go-redis/v9"
)

type RedisStorage struct {
	conn *redis.Client
}

func NewRedisStorage() KVStorage {
	client := redis.NewClient(&redis.Options{
		Addr:            "localhost:6379",
		DB:              0,
		WriteTimeout:    time.Second,
		ReadTimeout:     time.Second,
		PoolSize:        10,
		PoolTimeout:     time.Second,
		ConnMaxLifetime: time.Minute * 30,
		ConnMaxIdleTime: time.Minute,
	})

	return RedisStorage{
		conn: client,
	}
}
func (r RedisStorage) Save(ctx context.Context, key, value string) error {
	return r.conn.Set(ctx, key, value, 0).Err()
}

func (r RedisStorage) Get(ctx context.Context, key string) (string, error) {
	val, err := r.conn.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return "", pkg.ErrNotFound
		}

		return "", err
	}
	return val, nil
}
