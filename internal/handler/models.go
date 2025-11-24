package handler

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"openbridge/internal/config"
	"openbridge/internal/service"

	"github.com/gin-gonic/gin"
)

type ModelsHandler struct {
	config        *config.Config
	apiKeyManager *service.APIKeyManager
}

func NewModelsHandler(cfg *config.Config, apiKeyManager *service.APIKeyManager) *ModelsHandler {
	return &ModelsHandler{
		config:        cfg,
		apiKeyManager: apiKeyManager,
	}
}

// ListModels forwards the request to AssemblyAI and returns their model list
func (h *ModelsHandler) ListModels(c *gin.Context) {
	// Get API key
	apiKey := h.apiKeyManager.GetNextKey()
	if apiKey == "" {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "No API keys configured",
		})
		return
	}

	// Forward to AssemblyAI
	url := h.config.AssemblyAI.BaseURL + "/models"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Printf("Error creating request: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error forwarding to AssemblyAI: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading response: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Parse and return the response as-is
	var result interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		log.Printf("Error parsing response: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(resp.StatusCode, result)
}

// RetrieveModel forwards the request to AssemblyAI for a specific model
func (h *ModelsHandler) RetrieveModel(c *gin.Context) {
	modelID := c.Param("model")

	// Get API key
	apiKey := h.apiKeyManager.GetNextKey()
	if apiKey == "" {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "No API keys configured",
		})
		return
	}

	// Forward to AssemblyAI
	url := h.config.AssemblyAI.BaseURL + "/models/" + modelID
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Printf("Error creating request: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error forwarding to AssemblyAI: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading response: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Parse and return the response as-is
	var result interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		log.Printf("Error parsing response: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(resp.StatusCode, result)
}
