package admin

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"os"
	"sync"

	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v3"
)

// AdminConfig 管理后台的运行时配置
type AdminConfig struct {
	ClientAPIKeys []string                  `json:"client_api_keys" yaml:"client_api_keys"`
	Providers     map[string]ProviderConfig `json:"providers" yaml:"providers"`
	mu            sync.RWMutex
}

type ProviderConfig struct {
	Type             string   `json:"type" yaml:"type"`
	BaseURL          string   `json:"base_url" yaml:"base_url"`
	APIKeys          []string `json:"api_keys" yaml:"api_keys"`
	RotationStrategy string   `json:"rotation_strategy" yaml:"rotation_strategy"`
}

var (
	adminConfig *AdminConfig
	configPath  string
)

// Init 初始化管理配置
func Init(path string) error {
	configPath = path
	adminConfig = &AdminConfig{
		Providers: make(map[string]ProviderConfig),
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	return yaml.Unmarshal(data, adminConfig)
}

// SetupRoutes 设置管理后台路由
func SetupRoutes(r *gin.Engine, adminPassword string) {
	admin := r.Group("/admin")
	admin.Use(adminAuthMiddleware(adminPassword))
	{
		// 页面和资源
		admin.GET("", serveAdminPage)
		admin.GET("admin.js", serveAdminJS)

		// API
		admin.GET("/api/config", getConfig)
		admin.POST("/api/providers", addProvider)
		admin.DELETE("/api/providers/:name", deleteProvider)
		admin.POST("/api/keys/generate", generateClientKey)
		admin.DELETE("/api/keys/:key", deleteClientKey)
	}
}

func adminAuthMiddleware(password string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 简单的密码认证，从 query 或 header 获取
		auth := c.Query("password")
		if auth == "" {
			auth = c.GetHeader("X-Admin-Password")
		}

		if password != "" && auth != password {
			c.HTML(http.StatusUnauthorized, "", loginPage)
			c.Abort()
			return
		}
		c.Next()
	}
}

func getConfig(c *gin.Context) {
	adminConfig.mu.RLock()
	defer adminConfig.mu.RUnlock()

	// 隐藏 API Keys 的中间部分
	safeConfig := struct {
		ClientAPIKeys []string                  `json:"client_api_keys"`
		Providers     map[string]ProviderConfig `json:"providers"`
	}{
		ClientAPIKeys: adminConfig.ClientAPIKeys,
		Providers:     make(map[string]ProviderConfig),
	}

	for name, p := range adminConfig.Providers {
		maskedKeys := make([]string, len(p.APIKeys))
		for i, key := range p.APIKeys {
			maskedKeys[i] = maskKey(key)
		}
		safeConfig.Providers[name] = ProviderConfig{
			Type:             p.Type,
			BaseURL:          p.BaseURL,
			APIKeys:          maskedKeys,
			RotationStrategy: p.RotationStrategy,
		}
	}

	c.JSON(http.StatusOK, safeConfig)
}

func addProvider(c *gin.Context) {
	var req struct {
		Name             string   `json:"name"`
		Type             string   `json:"type"`
		BaseURL          string   `json:"base_url"`
		APIKeys          []string `json:"api_keys"`
		RotationStrategy string   `json:"rotation_strategy"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	adminConfig.mu.Lock()
	adminConfig.Providers[req.Name] = ProviderConfig{
		Type:             req.Type,
		BaseURL:          req.BaseURL,
		APIKeys:          req.APIKeys,
		RotationStrategy: req.RotationStrategy,
	}
	adminConfig.mu.Unlock()

	saveConfig()
	c.JSON(http.StatusOK, gin.H{"message": "Provider added", "name": req.Name})
}

func deleteProvider(c *gin.Context) {
	name := c.Param("name")

	adminConfig.mu.Lock()
	delete(adminConfig.Providers, name)
	adminConfig.mu.Unlock()

	saveConfig()
	c.JSON(http.StatusOK, gin.H{"message": "Provider deleted"})
}

func generateClientKey(c *gin.Context) {
	key := generateAPIKey()

	adminConfig.mu.Lock()
	adminConfig.ClientAPIKeys = append(adminConfig.ClientAPIKeys, key)
	adminConfig.mu.Unlock()

	saveConfig()
	c.JSON(http.StatusOK, gin.H{"key": key})
}

func deleteClientKey(c *gin.Context) {
	key := c.Param("key")

	adminConfig.mu.Lock()
	for i, k := range adminConfig.ClientAPIKeys {
		if k == key {
			adminConfig.ClientAPIKeys = append(adminConfig.ClientAPIKeys[:i], adminConfig.ClientAPIKeys[i+1:]...)
			break
		}
	}
	adminConfig.mu.Unlock()

	saveConfig()
	c.JSON(http.StatusOK, gin.H{"message": "Key deleted"})
}



func saveConfig() error {
	adminConfig.mu.RLock()
	defer adminConfig.mu.RUnlock()

	// 读取完整配置
	data, err := os.ReadFile(configPath)
	if err != nil {
		return err
	}

	var fullConfig map[string]interface{}
	if err := yaml.Unmarshal(data, &fullConfig); err != nil {
		return err
	}

	// 更新相关字段
	fullConfig["client_api_keys"] = adminConfig.ClientAPIKeys
	fullConfig["providers"] = adminConfig.Providers

	// 写回文件
	newData, err := yaml.Marshal(fullConfig)
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, newData, 0644)
}

func generateAPIKey() string {
	bytes := make([]byte, 24)
	rand.Read(bytes)
	return "sk-ob-" + hex.EncodeToString(bytes)
}

func maskKey(key string) string {
	if len(key) <= 8 {
		return "****"
	}
	return key[:4] + "****" + key[len(key)-4:]
}

func serveAdminPage(c *gin.Context) {
	c.File("internal/admin/admin.html")
}

func serveAdminJS(c *gin.Context) {
	c.File("internal/admin/admin.js")
}

var _ = json.Marshal // 确保 import 被使用

const loginPage = `<!DOCTYPE html>
<html>
<head><title>OpenBridge Admin</title></head>
<body style="display:flex;justify-content:center;align-items:center;height:100vh;font-family:system-ui">
<form method="GET">
<h2>🔐 OpenBridge Admin</h2>
<input name="password" type="password" placeholder="Password" style="padding:8px;margin-right:8px">
<button type="submit" style="padding:8px 16px">Login</button>
</form>
</body>
</html>`
