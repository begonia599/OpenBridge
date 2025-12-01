package provider

import (
	"fmt"
	"sync"
)

// Registry 管理所有注册的 Provider
type Registry struct {
	providers  map[string]Provider
	modelCache map[string]string // model ID -> provider name
	mu         sync.RWMutex
}

// NewRegistry 创建新的 Registry
func NewRegistry() *Registry {
	return &Registry{
		providers:  make(map[string]Provider),
		modelCache: make(map[string]string),
	}
}

// Register 注册一个 Provider
func (r *Registry) Register(name string, provider Provider) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.providers[name] = provider
}

// GetProvider 根据名称获取 Provider
func (r *Registry) GetProvider(name string) (Provider, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	p, ok := r.providers[name]
	return p, ok
}

// CacheModel 缓存模型到 Provider 的映射
func (r *Registry) CacheModel(modelID string, providerName string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.modelCache[modelID] = providerName
}

// RouteModel 根据模型名称路由到对应的 Provider
func (r *Registry) RouteModel(model string) (Provider, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// 从缓存中查找
	if providerName, ok := r.modelCache[model]; ok {
		if p, ok := r.providers[providerName]; ok {
			return p, nil
		}
	}

	// 如果只有一个 provider，直接返回
	if len(r.providers) == 1 {
		for _, p := range r.providers {
			return p, nil
		}
	}

	return nil, fmt.Errorf("no provider found for model: %s", model)
}

// ListProviders 列出所有注册的 Provider
func (r *Registry) ListProviders() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.providers))
	for name := range r.providers {
		names = append(names, name)
	}
	return names
}

// GetModelCache 获取模型缓存（调试用）
func (r *Registry) GetModelCache() map[string]string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	cache := make(map[string]string)
	for k, v := range r.modelCache {
		cache[k] = v
	}
	return cache
}
