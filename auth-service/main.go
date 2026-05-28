package main

import (
	"context"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"auth-service/config"
	"auth-service/pkg/handler"
	"auth-service/pkg/service"
	"auth-service/pkg/storage"
)

const version = "1.0.0"

func main() {
	// Load configuration
	cfg := config.Load()

	log.Printf("Starting Auth Service v%s", version)
	log.Printf("Configuration: Server Port=%s, Keycloak URL=%s, Realm=%s", 
		cfg.Server.Port, cfg.Keycloak.URL, cfg.Keycloak.Realm)

	// Initialize storage (Redis by default, with DB fallback)
	var store storage.Storage
	var err error

	if cfg.Database.Driver != "" && cfg.Database.DSN != "" {
		log.Printf("Using database storage: %s", cfg.Database.Driver)
		store, err = storage.NewDatabaseStorage(
			cfg.Database.Driver,
			cfg.Database.DSN,
			cfg.Database.MaxIdle,
			cfg.Database.MaxOpen,
		)
		if err != nil {
			log.Fatalf("Failed to initialize database storage: %v", err)
		}
	} else {
		log.Printf("Using Redis storage: %s", cfg.Redis.Addr)
		redisStore := storage.NewRedisStorage(cfg.Redis.Addr, cfg.Redis.Password, cfg.Redis.DB)
		
		// Test Redis connection
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		
		if err := redisStore.TestConnection(ctx); err != nil {
			log.Printf("Warning: Redis connection failed: %v. Running in degraded mode.", err)
		} else {
			log.Println("Redis connection successful")
		}
		
		store = redisStore
	}
	defer store.Close()

	// Initialize service
	authService := service.NewAuthService(cfg, store)

	// Initialize handler
	authHandler := handler.NewAuthHandler(authService)

	// Setup Gin router
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	// Add middleware for health check timestamp
	r.Use(func(c *gin.Context) {
		c.Set("time", time.Now().Format(time.RFC3339))
		c.Next()
	})

	// Register routes
	authHandler.RegisterRoutes(r)

	// Start server
	serverAddr := ":" + cfg.Server.Port
	log.Printf("Auth Service listening on %s", serverAddr)
	
	if err := r.Run(serverAddr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
