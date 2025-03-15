package kvstore

import (
	"kvmsg/util"
)

type impl struct {
	storage map[string][]([]byte)
}

func New() KVStore {
	return &impl{
		storage: make(map[string][]([]byte)),
	}
}

func CreateWithBackdoor() (KVStore, map[string][]([]byte)) {
	storage := make(map[string][]([]byte))
	return &impl{storage}, storage
}

func (im *impl) Put(key string, value []byte) {
	im.storage[key] = append(im.storage[key], value)
}

func (im *impl) Get(key string) []([]byte) {
	valueList := im.storage[key]
	return valueList
}

func (im *impl) Delete(key string) {
	delete(im.storage, key)
}

func (im *impl) Update(key string, oldVal []byte, newVal []byte) {
	valueList, exist := im.storage[key]
	if !exist {
		im.Put(key, newVal)
		return
	}

	markedIndex := util.FindIndex(valueList, oldVal)
	if markedIndex != -1 {
		im.Put(key, newVal)
		return
	}

	im.storage[key] = util.ReplaceElement(valueList, markedIndex, newVal)
}
