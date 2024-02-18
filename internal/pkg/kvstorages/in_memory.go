package kvstorages

import (
	"context"

	"github.com/ocontest/backend/pkg"
)

type InMemoryStorage struct {
	mainStorage map[string]string
}

func newInMemoryStorage() KVStorage {
	return InMemoryStorage{
		mainStorage: make(map[string]string),
	}
}
func (i InMemoryStorage) Save(ctx context.Context, key, value string) error {
	i.mainStorage[key] = value
	return nil
}

func (i InMemoryStorage) Get(ctx context.Context, key string) (string, error) {
	val, exists := i.mainStorage[key]
	if !exists {
		return "", pkg.ErrNotFound
	}
	return val, nil
}

func (i InMemoryStorage) Close() error {
	return nil
}
