package handler

import (
	"log"
	"net/http"
	"openbridge/internal/config"
	"openbridge/internal/models"
	"openbridge/internal/provider"
	"openbridge/internal/service"
	"sync"

	"github.com/gin-gonic/gin"
)

type ModelsHandler struct {
	config      *config.Config
	registry    *provider.Registry
	keyManagers *service.ProviderKeyManagers
}

func NewModelsHandler(cfg *config.Config, registry *provider.Registry, keyManagers *service.ProviderKeyManagers) *ModelsHandler {
	return &ModelsHandler{
		config:      cfg,
		registry:    registry,
		keyManagers: keyManagers,
	}
}

// ListModels 从所有上游 Provider 获取模型列表，并缓存映射关系
func (h *ModelsHandler) ListModels(c *gin.Context) {
	allModels := []models.Model{}
	var mu sync.Mutex
	var wg sync.WaitGroup

	// 并发从所有 Provider 获取模型
	for _, providerName := range h.registry.ListProviders() {
		wg.Add(1)
		go func(name string) {
			defer wg.Done()

			p, ok := h.registry.GetProvider(name)
			if !ok {
				return
			}

			apiKey := h.keyManagers.GetKey(name)
			if apiKey == "" {
				log.Printf("⚠️ No API key for provider: %s", name)
				return
			}

			modelList, err := p.ListModels(apiKey)
			if err != nil {
				log.Printf("⚠️ Failed to list models from %s: %v", name, err)
				return
			}

			mu.Lock()
			for _, m := range modelList.Data {
				// 缓存模型到 Provider 的映射
				h.registry.CacheModel(m.ID, name)
				// 标记模型来源
				m.OwnedBy = name
				allModels = append(allModels, m)
			}
			mu.Unlock()

			log.Printf("✅ Got %d models from %s", len(modelList.Data), name)
		}(providerName)
	}

	wg.Wait()

	c.JSON(http.StatusOK, models.ModelList{
		Object: "list",
		Data:   allModels,
	})
}

// RetrieveModel 获取单个模型信息
func (h *ModelsHandler) RetrieveModel(c *gin.Context) {
	modelID := c.Param("model")

	p, err := h.registry.RouteModel(modelID)
	if err != nil {
		c.JSON(http.StatusNotFound, models.NewErrorResponse(
			"Model not found: "+modelID,
			models.ErrorTypeNotFound,
			models.ErrorCodeModelNotFound,
		))
		return
	}

	c.JSON(http.StatusOK, models.Model{
		ID:      modelID,
		Object:  "model",
		OwnedBy: p.Name(),
	})
}
