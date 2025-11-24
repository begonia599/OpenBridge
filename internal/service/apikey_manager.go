package service

import (
	"math/rand"
	"sync"
	"sync/atomic"
)

type APIKeyManager struct {
	keys     []string
	strategy string

	// For round_robin
	currentIndex uint64

	// For least_used
	usageCount map[string]*uint64
	mu         sync.RWMutex
}

func NewAPIKeyManager(keys []string, strategy string) *APIKeyManager {
	manager := &APIKeyManager{
		keys:       keys,
		strategy:   strategy,
		usageCount: make(map[string]*uint64),
	}

	// Initialize usage counters
	for _, key := range keys {
		var count uint64
		manager.usageCount[key] = &count
	}

	return manager
}

func (m *APIKeyManager) GetNextKey() string {
	if len(m.keys) == 0 {
		return ""
	}

	switch m.strategy {
	case "random":
		return m.keys[rand.Intn(len(m.keys))]

	case "least_used":
		return m.getLeastUsedKey()

	case "round_robin":
		fallthrough
	default:
		return m.getRoundRobinKey()
	}
}

func (m *APIKeyManager) getRoundRobinKey() string {
	index := atomic.AddUint64(&m.currentIndex, 1) - 1
	key := m.keys[index%uint64(len(m.keys))]

	// Increment usage count
	if counter, ok := m.usageCount[key]; ok {
		atomic.AddUint64(counter, 1)
	}

	return key
}

func (m *APIKeyManager) getLeastUsedKey() string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var minKey string
	var minCount uint64 = ^uint64(0)

	for _, key := range m.keys {
		if counter, ok := m.usageCount[key]; ok {
			count := atomic.LoadUint64(counter)
			if count < minCount {
				minCount = count
				minKey = key
			}
		}
	}

	// Increment usage count
	if counter, ok := m.usageCount[minKey]; ok {
		atomic.AddUint64(counter, 1)
	}

	return minKey
}

func (m *APIKeyManager) GetStats() map[string]uint64 {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stats := make(map[string]uint64)
	for key, counter := range m.usageCount {
		stats[key] = atomic.LoadUint64(counter)
	}
	return stats
}
