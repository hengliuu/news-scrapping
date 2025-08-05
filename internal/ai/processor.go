package ai

import (
	"fmt"
	"log"

	"github.com/hengky/news-scrapping/internal/config"
	"github.com/hengky/news-scrapping/pkg/models"
)

// Processor handles the complete AI processing pipeline
type Processor struct {
	client *Client
	config *config.Config
}

// NewProcessor creates a new AI processor
func NewProcessor(cfg *config.Config) (*Processor, error) {
	client, err := New(cfg.GeminiAPIKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create AI client: %w", err)
	}

	return &Processor{
		client: client,
		config: cfg,
	}, nil
}

// Close closes the AI processor
func (p *Processor) Close() error {
	if p.client != nil {
		return p.client.Close()
	}
	return nil
}

// ProcessNewsItems processes scraped news and returns curated top 5 (backward compatibility)
func (p *Processor) ProcessNewsItems(newsItems []models.NewsItem) (*models.NewsResponse, error) {
	return p.ProcessNewsItemsByType(newsItems, "ai")
}

// ProcessNewsItemsByType processes scraped news based on type and returns curated top 5
func (p *Processor) ProcessNewsItemsByType(newsItems []models.NewsItem, newsType string) (*models.NewsResponse, error) {
	if len(newsItems) == 0 {
		log.Println("No news items to process")
		return &models.NewsResponse{News: []models.NewsItem{}}, nil
	}

	log.Printf("Processing %d %s news items with Gemini AI", len(newsItems), newsType)

	// Process with Gemini AI using type-specific processing
	response, err := p.client.ProcessNewsByType(newsItems, newsType)
	if err != nil {
		return nil, fmt.Errorf("failed to process news with AI: %w", err)
	}

	// Validate each news item in response
	var validNews []models.NewsItem
	for i, item := range response.News {
		if item.Title == "" {
			log.Printf("Warning: News item %d has empty title, skipping", i+1)
			continue
		}
		if item.URL == "" {
			log.Printf("Warning: News item %d has empty URL, skipping", i+1)
			continue
		}
		if item.Summary == "" {
			log.Printf("Warning: News item %d has empty summary, using title", i+1)
			item.Summary = item.Title
		}
		if item.Source == "" {
			item.Source = "Unknown"
		}

		validNews = append(validNews, item)
	}

	response.News = validNews
	
	log.Printf("AI processing completed: %d valid %s news items selected", len(response.News), newsType)

	return response, nil
}