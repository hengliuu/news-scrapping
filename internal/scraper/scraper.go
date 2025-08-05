package scraper

import (
	"fmt"
	"log"
	"sync"

	"github.com/hengky/news-scrapping/pkg/models"
)

// Scraper handles news scraping operations
type Scraper struct {
	aiSources     []NewsSource
	globalSources []NewsSource
}

// New creates a new scraper instance
func New() *Scraper {
	return &Scraper{
		aiSources:     GetAINewsSources(),
		globalSources: GetGlobalNewsSources(),
	}
}

// ScrapeAllSources scrapes news from all AI sources (backward compatibility)
func (s *Scraper) ScrapeAllSources() ([]models.NewsItem, error) {
	return s.ScrapeNewsByType("ai")
}

// ScrapeNewsByType scrapes news from sources based on type
func (s *Scraper) ScrapeNewsByType(newsType string) ([]models.NewsItem, error) {
	var sources []NewsSource
	
	// Select sources based on type
	switch newsType {
	case "global":
		sources = s.globalSources
	case "ai":
		fallthrough
	default:
		sources = s.aiSources
		newsType = "ai" // Normalize the type
	}

	var allNews []models.NewsItem
	var mu sync.Mutex
	var wg sync.WaitGroup

	// Channel to collect errors
	errChan := make(chan error, len(sources))

	// Scrape from all sources concurrently
	for _, source := range sources {
		wg.Add(1)
		go func(src NewsSource) {
			defer wg.Done()

			news, err := ScrapeNewsFromSource(src, newsType)
			if err != nil {
				log.Printf("Error scraping from %s: %v", src.Name, err)
				errChan <- fmt.Errorf("failed to scrape %s: %w", src.Name, err)
				return
			}

			// Thread-safe append
			mu.Lock()
			allNews = append(allNews, news...)
			mu.Unlock()

			log.Printf("Successfully scraped %d articles from %s", len(news), src.Name)
		}(source)
	}

	// Wait for all goroutines to complete
	wg.Wait()
	close(errChan)

	// Collect any errors (but don't fail if some sources work)
	var errors []error
	for err := range errChan {
		errors = append(errors, err)
	}

	if len(allNews) == 0 {
		return nil, fmt.Errorf("no news articles scraped from any source. Errors: %v", errors)
	}

	// Log summary
	log.Printf("Total scraped %s articles: %d from %d sources", newsType, len(allNews), len(sources))
	if len(errors) > 0 {
		log.Printf("Encountered %d errors during scraping", len(errors))
	}

	return allNews, nil
}

// GetSourceCount returns the number of AI sources (backward compatibility)
func (s *Scraper) GetSourceCount() int {
	return len(s.aiSources)
}

// GetSourceCountByType returns the number of sources for a specific type
func (s *Scraper) GetSourceCountByType(newsType string) int {
	switch newsType {
	case "global":
		return len(s.globalSources)
	case "ai":
		fallthrough
	default:
		return len(s.aiSources)
	}
}