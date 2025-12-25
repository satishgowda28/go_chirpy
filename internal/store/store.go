package store

import "sync"

var (
	m     sync.RWMutex
	store = make(map[string]interface{})
)

func SetValue(key string, value interface{}) {
	m.Lock()
	defer m.Unlock()
	store[key] = value
}

func GetValue(key string) interface{} {
	m.RLock()
	defer m.RUnlock()
	return store[key]
}
