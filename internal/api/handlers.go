package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hengky/news-scrapping/internal/ai"
	"github.com/hengky/news-scrapping/internal/config"
	"github.com/hengky/news-scrapping/internal/discord"
	"github.com/hengky/news-scrapping/internal/scheduler"
	"github.com/hengky/news-scrapping/internal/scraper"
	"github.com/hengky/news-scrapping/pkg/models"
)

// Handlers contains all HTTP handlers
type Handlers struct {
	config    *config.Config
	scheduler *scheduler.Scheduler
}

// NewHandlers creates a new handlers instance
func NewHandlers(cfg *config.Config) *Handlers {
	// Create scheduler for manual operations
	schedulerInstance := scheduler.New(cfg)

	return &Handlers{
		config:    cfg,
		scheduler: schedulerInstance,
	}
}

// HealthCheck returns service health status
func (h *Handlers) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"version":   "1.0.0",
		"timestamp": time.Now().UTC(),
		"service":   "news-scrapping-service",
	})
}

// RootHandler handles root path requests
func (h *Handlers) RootHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "AI Tech News Scrapping Service",
		"version": "1.0.0",
		"endpoints": gin.H{
			"health":  "/health",
			"status":  "/api/v1/status",
			"trigger": "/api/v1/trigger (POST)",
			"latest":  "/api/v1/latest",
		},
	})
}

// GetStatus returns the current job status
func (h *Handlers) GetStatus(c *gin.Context) {
	status := h.scheduler.GetJobStatus()
	c.JSON(http.StatusOK, status)
}

// TriggerNews manually triggers news scraping
func (h *Handlers) TriggerNews(c *gin.Context) {
	// Get news type from query parameter
	newsType := c.DefaultQuery("type", "ai")
	if newsType != "ai" && newsType != "global" {
		newsType = "ai" // Default to AI for invalid types
	}

	// Check if job is already running
	if h.scheduler.IsRunning() {
		c.JSON(http.StatusTooManyRequests, models.APIResponse{
			Message: "News job is already running",
			Error:   "Job in progress",
		})
		return
	}

	// Run job in background with specified type
	go func() {
		if err := h.scheduler.RunManualJobByType(newsType); err != nil {
			// Log error but don't fail the API response
			// since the job runs asynchronously
			// TODO: Add proper logging
		}
	}()

	message := fmt.Sprintf("%s news scraping job triggered successfully", newsType)
	if newsType == "ai" {
		message = "AI tech news scraping job triggered successfully"
	} else {
		message = "Global tech news scraping job triggered successfully"
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Message: message,
		Data: gin.H{
			"triggered_at": time.Now().UTC(),
			"type":         newsType,
		},
	})
}

// GetLatestNews gets the latest news without sending to Discord
func (h *Handlers) GetLatestNews(c *gin.Context) {
	// Get news type from query parameter
	newsType := c.DefaultQuery("type", "ai")
	if newsType != "ai" && newsType != "global" {
		newsType = "ai" // Default to AI for invalid types
	}

	// Create temporary instances for this request
	scraperInstance := scraper.New()
	
	aiProcessor, err := ai.NewProcessor(h.config)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Message: "Failed to initialize AI processor",
			Error:   err.Error(),
		})
		return
	}
	defer aiProcessor.Close()

	// Step 1: Scrape news with specified type
	newsItems, err := scraperInstance.ScrapeNewsByType(newsType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Message: "Failed to scrape news",
			Error:   err.Error(),
		})
		return
	}

	if len(newsItems) == 0 {
		c.JSON(http.StatusOK, models.APIResponse{
			Message: "No news items found",
			Data: gin.H{
				"news":         []models.NewsItem{},
				"generated_at": time.Now().UTC(),
				"source_count": scraperInstance.GetSourceCount(),
			},
		})
		return
	}

	// Step 2: Process with AI using specified type
	newsResponse, err := aiProcessor.ProcessNewsItemsByType(newsItems, newsType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Message: "Failed to process news with AI",
			Error:   err.Error(),
		})
		return
	}

	// Build response data
	responseData := gin.H{
		"news":           newsResponse.News,
		"generated_at":   time.Now().UTC(),
		"scraped_count":  len(newsItems),
		"selected_count": len(newsResponse.News),
		"source_count":   scraperInstance.GetSourceCountByType(newsType),
		"type":          newsType,
	}

	// Add token usage if available
	if newsResponse.TokenUsage != nil {
		responseData["token_usage"] = newsResponse.TokenUsage
	}

	// Return processed news
	message := fmt.Sprintf("Latest %s news retrieved successfully", newsType)
	if newsType == "ai" {
		message = "Latest AI tech news retrieved successfully"
	} else {
		message = "Latest global tech news retrieved successfully"
	}
	
	c.JSON(http.StatusOK, models.APIResponse{
		Message: message,
		Data:    responseData,
	})
}

// TestDiscord tests Discord webhook (utility endpoint)
func (h *Handlers) TestDiscord(c *gin.Context) {
	discordClient := discord.New(h.config.DiscordWebhook)
	
	if err := discordClient.TestWebhook(); err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Message: "Discord webhook test failed",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Message: "Discord webhook test successful",
		Data: gin.H{
			"tested_at": time.Now().UTC(),
		},
	})
}