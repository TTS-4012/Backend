package kvstorages

import (
	"context"
	"errors"

	"github.com/ocontest/backend/pkg/configs"
)

type KVStorage interface {
	Save(ctx context.Context, key, value string) error
	Get(ctx context.Context, key string) (string, error)
}

func NewKVStorage(c configs.SectionKVStore) (KVStorage, error) {
	switch c.Type {
	case "redis":
		return newRedisStorage(c.Redis)
	case "in_memory":
		return newInMemoryStorage(), nil
	default:
		return nil, errors.New("Unknown kvstore type")
	}

}
