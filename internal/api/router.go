package api

import (
	"github.com/gin-gonic/gin"
	"github.com/hengky/news-scrapping/internal/config"
)

// SetupRouter sets up the Gin router with all routes
func SetupRouter(cfg *config.Config) *gin.Engine {
	// Set Gin mode
	gin.SetMode(cfg.GinMode)

	// Create router
	router := gin.New()

	// Add middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(corsMiddleware())

	// Create handlers
	handlers := NewHandlers(cfg)

	// Routes
	v1 := router.Group("/api/v1")
	{
		v1.GET("/health", handlers.HealthCheck)
		v1.GET("/status", handlers.GetStatus)
		v1.POST("/trigger", handlers.TriggerNews)
		v1.GET("/latest", handlers.GetLatestNews)
	}

	// Root health check
	router.GET("/health", handlers.HealthCheck)
	router.GET("/", handlers.RootHandler)

	return router
}