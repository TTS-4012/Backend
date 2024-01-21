package kvstorages

import (
	"context"
	"fmt"

	"github.com/ocontest/backend/pkg"
	"github.com/ocontest/backend/pkg/configs"
	"github.com/redis/go-redis/v9"
)

type RedisStorage struct {
	conn *redis.Client
}

func newRedisStorage(c configs.SectionRedis) (KVStorage, error) {
	fmt.Println(c, "redis config")
	client := redis.NewClient(&redis.Options{
		Addr:         c.Address,
		DB:           c.DB,
		WriteTimeout: c.Timeout,
		ReadTimeout:  c.Timeout,
		PoolSize:     10,
		PoolTimeout:  c.Timeout,
	})

	ctx, cancel := context.WithTimeout(context.Background(), c.Timeout)
	defer cancel()
	err := client.Ping(ctx).Err()

	return RedisStorage{
		conn: client,
	}, err
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
