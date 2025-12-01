package service

import (
	"math/rand"
	"sync"
	"sync/atomic"
)

// APIKeyManager 管理单个 Provider 的 API Keys
type APIKeyManager struct {
	keys     []string
	strategy string

	// For round_robin
	currentIndex uint64

	// For least_used
	usageCount map[string]*uint64
	mu         sync.RWMutex
}

// NewAPIKeyManager 创建新的 APIKeyManager
func NewAPIKeyManager(keys []string, strategy string) *APIKeyManager {
	if strategy == "" {
		strategy = "round_robin"
	}

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

// GetNextKey 获取下一个 API Key
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

// GetStats 获取使用统计
func (m *APIKeyManager) GetStats() map[string]uint64 {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stats := make(map[string]uint64)
	for key, counter := range m.usageCount {
		// Mask key for security
		maskedKey := maskKey(key)
		stats[maskedKey] = atomic.LoadUint64(counter)
	}
	return stats
}

// maskKey 隐藏 API Key 中间部分
func maskKey(key string) string {
	if len(key) <= 8 {
		return "****"
	}
	return key[:4] + "****" + key[len(key)-4:]
}

// ProviderKeyManagers 管理所有 Provider 的 API Keys
type ProviderKeyManagers struct {
	managers map[string]*APIKeyManager
	mu       sync.RWMutex
}

// NewProviderKeyManagers 创建新的 ProviderKeyManagers
func NewProviderKeyManagers() *ProviderKeyManagers {
	return &ProviderKeyManagers{
		managers: make(map[string]*APIKeyManager),
	}
}

// Register 注册一个 Provider 的 APIKeyManager
func (p *ProviderKeyManagers) Register(providerName string, keys []string, strategy string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.managers[providerName] = NewAPIKeyManager(keys, strategy)
}

// GetKey 获取指定 Provider 的下一个 API Key
func (p *ProviderKeyManagers) GetKey(providerName string) string {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if manager, ok := p.managers[providerName]; ok {
		return manager.GetNextKey()
	}
	return ""
}

// GetStats 获取所有 Provider 的使用统计
func (p *ProviderKeyManagers) GetStats() map[string]map[string]uint64 {
	p.mu.RLock()
	defer p.mu.RUnlock()

	stats := make(map[string]map[string]uint64)
	for name, manager := range p.managers {
		stats[name] = manager.GetStats()
	}
	return stats
}
