package kvstorages

type KVStorage interface {
	Save(key, value string) error
	Get(key string) (string, error)
}
