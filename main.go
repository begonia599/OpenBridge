package main

import (
	"log"
	"openbridge/internal/config"
	"openbridge/internal/router"
	"openbridge/internal/service"
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

	// Initialize API key manager
	apiKeyManager := service.NewAPIKeyManager(cfg.AssemblyAI.APIKeys, cfg.AssemblyAI.RotationStrategy)

	// Setup router
	r := router.Setup(cfg, apiKeyManager)

	// Start server
	addr := cfg.Server.Host + ":" + cfg.Server.Port
	log.Printf("🌐 OpenBridge v%s starting on %s", Version, addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
