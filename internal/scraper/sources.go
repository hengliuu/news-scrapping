package scraper

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/hengky/news-scrapping/pkg/models"
	"github.com/mmcdole/gofeed"
)

// NewsSource represents a news source configuration
type NewsSource struct {
	Name string
	URL  string
	Type string // "rss" or "web"
}

// GetAINewsSources returns AI-focused tech news sources
func GetAINewsSources() []NewsSource {
	return []NewsSource{
		{
			Name: "TechCrunch AI",
			URL:  "https://techcrunch.com/category/artificial-intelligence/feed/",
			Type: "rss",
		},
		{
			Name: "The Verge AI",
			URL:  "https://www.theverge.com/ai-artificial-intelligence/rss/index.xml",
			Type: "rss",
		},
		{
			Name: "AI News",
			URL:  "https://artificialintelligence-news.com/feed/",
			Type: "rss",
		},
		{
			Name: "Bloomberg Technology",
			URL:  "https://feeds.bloomberg.com/technology/news.rss",
			Type: "rss",
		},
	}
}

// GetGlobalNewsSources returns global tech/business news sources
func GetGlobalNewsSources() []NewsSource {
	return []NewsSource{
		{
			Name: "Bloomberg Technology",
			URL:  "https://feeds.bloomberg.com/technology/news.rss",
			Type: "rss",
		},
		{
			Name: "Bloomberg Crypto",
			URL:  "https://feeds.bloomberg.com/crypto/news.rss",
			Type: "rss",
		},
		{
			Name: "Bloomberg Politics",
			URL:  "https://feeds.bloomberg.com/politics/news.rss",
			Type: "rss",
		},
		{
			Name: "Bloomberg Economics",
			URL:  "https://feeds.bloomberg.com/politics/news.rss",
			Type: "rss",
		},
	}
}

// GetNewsSourcesByType returns news sources based on type
func GetNewsSourcesByType(newsType string) []NewsSource {
	switch newsType {
	case "global":
		return GetGlobalNewsSources()
	case "ai":
		fallthrough
	default:
		return GetAINewsSources()
	}
}

// ScrapeNewsFromSource scrapes news from a single source with type filtering
func ScrapeNewsFromSource(source NewsSource, newsType string) ([]models.NewsItem, error) {
	switch source.Type {
	case "rss":
		return scrapeRSSFeed(source, newsType)
	default:
		return nil, fmt.Errorf("unsupported source type: %s", source.Type)
	}
}

// scrapeRSSFeed scrapes news from RSS feed
func scrapeRSSFeed(source NewsSource, newsType string) ([]models.NewsItem, error) {
	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// Create feed parser
	fp := gofeed.NewParser()
	fp.Client = client

	// Parse the feed
	feed, err := fp.ParseURL(source.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse RSS feed from %s: %w", source.Name, err)
	}

	var newsItems []models.NewsItem

	// Process recent items (last 24 hours)
	cutoff := time.Now().Add(-24 * time.Hour)

	for _, item := range feed.Items {
		// Parse published date
		var publishedAt time.Time
		if item.PublishedParsed != nil {
			publishedAt = *item.PublishedParsed
		} else if item.UpdatedParsed != nil {
			publishedAt = *item.UpdatedParsed
		} else {
			publishedAt = time.Now()
		}

		// Only include recent items
		if publishedAt.Before(cutoff) {
			continue
		}

		// Apply content filtering based on news type
		shouldInclude := true
		if newsType == "ai" {
			// For AI news, filter for AI-related content
			shouldInclude = isAIRelated(item.Title + " " + item.Description)
		} else if newsType == "global" {
			// For global news, use more specific filtering for markets/business/crypto
			shouldInclude = isGlobalBusinessRelated(item.Title + " " + item.Description)
		}

		if !shouldInclude {
			continue
		}

		// Create news item
		newsItem := models.NewsItem{
			Title:       cleanText(item.Title),
			Summary:     cleanText(item.Description),
			URL:         item.Link,
			Source:      source.Name,
			PublishedAt: publishedAt,
		}

		// Limit summary length
		if len(newsItem.Summary) > 300 {
			newsItem.Summary = newsItem.Summary[:297] + "..."
		}

		newsItems = append(newsItems, newsItem)

		// Limit items per source based on type
		maxItemsPerSource := 12
		if newsType == "global" {
			maxItemsPerSource = 8 // Reasonable limit for global news
		}
		
		if len(newsItems) >= maxItemsPerSource {
			break
		}
	}

	log.Printf("Scraped %d %s articles from %s", len(newsItems), newsType, source.Name)
	return newsItems, nil
}

// isAIRelated checks if the content is AI-related
func isAIRelated(content string) bool {
	content = strings.ToLower(content)

	aiKeywords := []string{
		"artificial intelligence", "ai", "machine learning", "ml",
		"deep learning", "neural network", "chatgpt", "openai",
		"google ai", "microsoft ai", "anthropic", "claude",
		"generative ai", "llm", "large language model",
		"computer vision", "natural language processing", "nlp",
		"automation", "robotics", "algorithm", "data science",
		"tensorflow", "pytorch", "hugging face",
	}

	for _, keyword := range aiKeywords {
		if strings.Contains(content, keyword) {
			return true
		}
	}

	return false
}

// isTechRelated checks if the content is tech-related (broader than AI)
func isTechRelated(content string) bool {
	content = strings.ToLower(content)

	techKeywords := []string{
		// Core tech
		"technology", "tech", "software", "hardware", "computing", "digital",
		"internet", "web", "mobile", "app", "platform", "cloud", "data",
		"cyber", "security", "privacy", "blockchain", "crypto", "fintech",

		// AI and related
		"artificial intelligence", "ai", "machine learning", "ml", "automation",
		"neural network", "algorithm", "chatgpt", "openai", "google ai",

		// Business tech
		"startup", "venture", "funding", "ipo", "acquisition", "merger",
		"innovation", "disruption", "digital transformation", "saas",

		// Companies and trends
		"microsoft", "apple", "google", "amazon", "meta", "tesla", "nvidia",
		"semiconductor", "chip", "processor", "quantum", "5g", "iot",
		"virtual reality", "augmented reality", "metaverse", "gaming",

		// Programming and development
		"programming", "developer", "coding", "api", "open source",
		"github", "database", "analytics", "big data",
	}

	for _, keyword := range techKeywords {
		if strings.Contains(content, keyword) {
			return true
		}
	}

	return false
}

// isGlobalBusinessRelated checks if content is relevant for global business/markets/crypto news
func isGlobalBusinessRelated(content string) bool {
	content = strings.ToLower(content)
	
	// More specific keywords for global business focus
	globalKeywords := []string{
		// Markets & Economics
		"stock", "market", "nasdaq", "dow", "s&p", "earnings", "revenue", "profit",
		"economy", "inflation", "recession", "gdp", "interest rate", "federal reserve",
		"economic", "fiscal", "monetary", "trade war", "tariff",
		
		// Crypto & Finance
		"bitcoin", "ethereum", "crypto", "cryptocurrency", "blockchain", "defi",
		"fintech", "banking", "payment", "financial services", "lending",
		
		// Business & Corporate
		"merger", "acquisition", "ipo", "funding", "investment", "venture capital",
		"startup", "unicorn", "valuation", "ceo", "executive", "leadership",
		"corporate", "business strategy", "partnerships",
		
		// Politics & Policy (tech related)
		"regulation", "policy", "government", "senate", "congress", "biden",
		"trump", "china", "trade", "sanctions", "antitrust", "monopoly",
		
		// Major Companies (when in business context)
		"apple", "microsoft", "google", "amazon", "meta", "tesla", "nvidia",
		"samsung", "tsmc", "intel", "amd",
	}

	for _, keyword := range globalKeywords {
		if strings.Contains(content, keyword) {
			return true
		}
	}

	return false
}

// cleanText removes HTML tags and extra whitespace
func cleanText(text string) string {
	// Simple HTML tag removal
	text = strings.ReplaceAll(text, "<p>", "")
	text = strings.ReplaceAll(text, "</p>", " ")
	text = strings.ReplaceAll(text, "<br>", " ")
	text = strings.ReplaceAll(text, "<br/>", " ")
	text = strings.ReplaceAll(text, "<br />", " ")

	// Remove other common HTML tags
	htmlTags := []string{"<div>", "</div>", "<span>", "</span>", "<a>", "</a>",
		"<strong>", "</strong>", "<b>", "</b>", "<i>", "</i>",
		"<em>", "</em>", "<u>", "</u>"}

	for _, tag := range htmlTags {
		text = strings.ReplaceAll(text, tag, "")
	}

	// Clean up whitespace
	text = strings.TrimSpace(text)
	text = strings.ReplaceAll(text, "\n", " ")
	text = strings.ReplaceAll(text, "\t", " ")

	// Remove multiple spaces
	for strings.Contains(text, "  ") {
		text = strings.ReplaceAll(text, "  ", " ")
	}

	return text
}
