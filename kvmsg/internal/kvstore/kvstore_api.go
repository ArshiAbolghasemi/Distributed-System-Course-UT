package kvstore

type KVStore interface {
	Put(key string, value []byte)
	Get(key string) []([]byte)
	Delete(key string)
	Update(key string, oldVal []byte, newVal []byte)
}

