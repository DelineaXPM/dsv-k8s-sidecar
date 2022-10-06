package cache

import (
	log "github.com/sirupsen/logrus"
	"sync"
)

type Cache interface {
	Get(key string) interface{}
	KeySet() []string
	Set(key string, value interface{})
}

type memoryCache struct {
	valueMap map[string]interface{}
	mu       sync.RWMutex
}

func CreateMemoryCache() Cache {
	return &memoryCache{
		valueMap: make(map[string]interface{}),
	}
}

func (m *memoryCache) Get(key string) interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()
	log.WithField("key", key).Info("Cache lookup")
	return m.valueMap[key]
}

func (m *memoryCache) Set(key string, value interface{}) {
	m.mu.Lock()
	m.valueMap[key] = value
	m.mu.Unlock()
}

func (m *memoryCache) KeySet() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	keys := make([]string, 0, len(m.valueMap))
	for k := range m.valueMap {
		keys = append(keys, k)
	}
	return keys
}
