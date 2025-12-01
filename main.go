package main

import (
	"log"
	"openbridge/internal/admin"
	"openbridge/internal/config"
	"openbridge/internal/provider"
	"openbridge/internal/provider/anthropic"
	"openbridge/internal/provider/google"
	"openbridge/internal/provider/openai"
	"openbridge/internal/router"
	"openbridge/internal/service"
	"openbridge/internal/user"
)

func main() {
	// Print version banner
	log.Println("========================================")
	log.Printf("🚀 OpenBridge v%s", Version)
	log.Printf("📅 Build Date: %s", BuildDate)
	log.Printf("📝 %s", Description)
	log.Println("========================================")
	log.Println()

	// Load configuration
	cfg, err := config.Load("config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize provider registry
	registry := provider.NewRegistry()

	// Initialize API key managers
	keyManagers := service.NewProviderKeyManagers()

	// Register providers from config
	for name, providerCfg := range cfg.Providers {
		var p provider.Provider

		switch providerCfg.Type {
		case "openai":
			p = openai.New(name, providerCfg.BaseURL)
		case "anthropic", "claude":
			p = anthropic.New(name, providerCfg.BaseURL)
		case "google", "gemini":
			p = google.New(name, providerCfg.BaseURL)
		default:
			// 默认使用 OpenAI 格式
			log.Printf("⚠️  Unknown provider type '%s', using OpenAI format", providerCfg.Type)
			p = openai.New(name, providerCfg.BaseURL)
		}

		registry.Register(name, p)
		keyManagers.Register(name, providerCfg.APIKeys, providerCfg.RotationStrategy)
		
		baseURL := providerCfg.BaseURL
		if baseURL == "" {
			baseURL = "(default)"
		}
		log.Printf("✅ Registered provider: %s (%s) -> %s", name, providerCfg.Type, baseURL)
	}

	// Setup router
	r := router.Setup(cfg, registry, keyManagers)

	// Setup user system
	if err := user.Init("users.json"); err != nil {
		log.Printf("⚠️ Failed to init user system: %v", err)
	} else {
		user.SetupRoutes(r)
		log.Printf("👤 User system enabled at /user")
	}

	// Setup admin panel
	if cfg.Admin.Enabled {
		if err := admin.Init("config.yaml"); err != nil {
			log.Printf("⚠️ Failed to init admin: %v", err)
		} else {
			admin.SetupRoutes(r, cfg.Admin.Password)
			log.Printf("🔧 Admin panel enabled at /admin")
		}
	}

	// Start server
	addr := cfg.Server.Host + ":" + cfg.Server.Port
	log.Printf("🌐 OpenBridge v%s starting on %s", Version, addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
