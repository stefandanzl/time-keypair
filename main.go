package main

import (
	"data-cron-server/api"
	"data-cron-server/config"
	"data-cron-server/cron"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

func main() {
	// Load environment variables or use defaults
	port := getEnvOrDefault("PORT", "8080")
	superAdminKey := getEnvOrDefault("SUPER_ADMIN_KEY", "super_admin_key")
	configFilePath := getEnvOrDefault("CONFIG_FILE_PATH", "/config/config.json")
	autoSaveIntervalStr := getEnvOrDefault("AUTO_SAVE_INTERVAL", "60")

	autoSaveInterval, err := strconv.Atoi(autoSaveIntervalStr)
	if err != nil {
		log.Fatalf("Invalid AUTO_SAVE_INTERVAL: %v", err)
	}

	// Initialize configuration
	cfg, err := config.LoadConfig(configFilePath)
	if err != nil {
		log.Printf("Starting with empty configuration: %v", err)
		cfg = config.NewConfig()
	}

	// Initialize cron scheduler
	scheduler := cron.NewScheduler(cfg)

	// Start auto-save goroutine
	stopChan := make(chan struct{})
	go autoSaveConfig(cfg, configFilePath, time.Duration(autoSaveInterval)*time.Second, stopChan)

	// Initialize API router
	router := api.NewRouter(cfg, scheduler, superAdminKey)

	// Start HTTP server
	server := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Shutting down server...")

		// Save configuration one last time
		if err := config.SaveConfig(cfg, configFilePath); err != nil {
			log.Printf("Error saving config on shutdown: %v", err)
		}

		// Stop auto-save goroutine
		close(stopChan)

		// Stop scheduler
		scheduler.Stop()

		// Shutdown server
		if err := server.Close(); err != nil {
			log.Printf("Error shutting down server: %v", err)
		}
	}()

	log.Printf("Server starting on port %s", port)
	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatalf("Server failed: %v", err)
	}
}

func getEnvOrDefault(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func autoSaveConfig(cfg *config.Config, filePath string, interval time.Duration, stopChan <-chan struct{}) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := config.SaveConfig(cfg, filePath); err != nil {
				log.Printf("Error auto-saving config: %v", err)
			} else {
				log.Printf("Configuration auto-saved to %s", filePath)
			}
		case <-stopChan:
			log.Println("Auto-save stopped")
			return
		}
	}
}
