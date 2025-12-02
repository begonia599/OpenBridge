package provider

import (
	"fmt"
	"sync"
)

// Registry 管理所有注册的 Provider
type Registry struct {
	providers  map[string]Provider
	modelCache map[string]ModelCacheEntry // prefixed model ID -> cache entry
	mu         sync.RWMutex
}

// ModelCacheEntry 缓存条目，存储 Provider 名称和实际模型 ID
type ModelCacheEntry struct {
	ProviderName string // Provider 的名称
	ActualModel  string // 实际的模型 ID（不带前缀）
}

// NewRegistry 创建新的 Registry
func NewRegistry() *Registry {
	return &Registry{
		providers:  make(map[string]Provider),
		modelCache: make(map[string]ModelCacheEntry),
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
// prefixedID: "provider_name/model_id"
// providerName: Provider 的名称
// actualModel: 实际的模型 ID（不带前缀）
func (r *Registry) CacheModel(prefixedID string, providerName string, actualModel string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.modelCache[prefixedID] = ModelCacheEntry{
		ProviderName: providerName,
		ActualModel:  actualModel,
	}
}

// RouteModel 根据模型名称路由到对应的 Provider
// 支持两种格式：
// 1. "provider_name/model_id" - 带前缀的格式
// 2. "model_id" - 不带前缀的格式（仅当只有一个 Provider 时）
// 返回: (providerName, actualModel, error)
func (r *Registry) RouteModel(model string) (string, string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// 检查是否是带前缀的格式 "provider_name/model_id"
	if entry, ok := r.modelCache[model]; ok {
		if _, providerExists := r.providers[entry.ProviderName]; providerExists {
			return entry.ProviderName, entry.ActualModel, nil
		}
	}

	// 如果只有一个 provider，直接返回（向后兼容）
	if len(r.providers) == 1 {
		for name := range r.providers {
			return name, model, nil
		}
	}

	return "", "", fmt.Errorf("no provider found for model: %s", model)
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
func (r *Registry) GetModelCache() map[string]ModelCacheEntry {
	r.mu.RLock()
	defer r.mu.RUnlock()

	cache := make(map[string]ModelCacheEntry)
	for k, v := range r.modelCache {
		cache[k] = v
	}
	return cache
}
