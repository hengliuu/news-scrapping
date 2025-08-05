package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/google/generative-ai-go/genai"
	"github.com/hengky/news-scrapping/pkg/models"
	"google.golang.org/api/option"
)

// Client handles Gemini AI operations
type Client struct {
	client *genai.Client
	model  *genai.GenerativeModel
}

// New creates a new Gemini AI client
func New(apiKey string) (*Client, error) {
	ctx := context.Background()
	
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini client: %w", err)
	}

	// Use gemini-2.5-flash model as specified
	model := client.GenerativeModel("gemini-2.5-flash")
	
	// Configure model parameters - use default max output tokens
	model.SetTemperature(0.3)
	model.SetTopK(40)
	model.SetTopP(0.95)
	// Using default max output tokens (1,048,576) for gemini-2.5-flash

	return &Client{
		client: client,
		model:  model,
	}, nil
}

// Close closes the Gemini client
func (c *Client) Close() error {
	return c.client.Close()
}

// ProcessNews processes scraped news and returns top 5 AI tech news (backward compatibility)
func (c *Client) ProcessNews(newsItems []models.NewsItem) (*models.NewsResponse, error) {
	return c.ProcessNewsByType(newsItems, "ai")
}

// ProcessNewsByType processes scraped news based on type and returns top 5
func (c *Client) ProcessNewsByType(newsItems []models.NewsItem, newsType string) (*models.NewsResponse, error) {
	if len(newsItems) == 0 {
		return &models.NewsResponse{News: []models.NewsItem{}}, nil
	}

	ctx := context.Background()

	// Limit news items to prevent overwhelming the AI
	var maxArticles int
	if newsType == "global" {
		maxArticles = 15 // Global news can handle more with default token limit
	} else {
		maxArticles = 12 // AI news
	}
	
	if len(newsItems) > maxArticles {
		log.Printf("Limiting news items from %d to %d to prevent token overflow", len(newsItems), maxArticles)
		newsItems = newsItems[:maxArticles]
	}

	// Limit summary length for better processing
	for i := range newsItems {
		if len(newsItems[i].Summary) > 200 {
			newsItems[i].Summary = newsItems[i].Summary[:197] + "..."
		}
	}

	// Convert news items to JSON for the prompt
	articlesJSON, err := json.MarshalIndent(newsItems, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal news items: %w", err)
	}

	// Estimate token count (rough approximation: 1 token ≈ 4 characters)
	estimatedTokens := len(string(articlesJSON)) / 4
	log.Printf("Estimated input tokens: %d (from %d articles)", estimatedTokens, len(newsItems))

	// Create type-specific prompt - simplified to reduce output tokens
	var prompt string
	if newsType == "global" {
		prompt = fmt.Sprintf(`Select TOP 5 global business/tech/crypto news:

%s

JSON:
{"news":[{"title":"","summary":"","url":"","source":"","relevance":""}]}`, string(articlesJSON))
	} else {
		// AI-focused prompt (default)
		prompt = fmt.Sprintf(`Select TOP 5 AI tech news:

%s

JSON:
{"news":[{"title":"","summary":"","url":"","source":"","relevance":""}]}`, string(articlesJSON))
	}

	// Generate content
	resp, err := c.model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return nil, fmt.Errorf("failed to generate content: %w", err)
	}

	if len(resp.Candidates) == 0 {
		return nil, fmt.Errorf("no candidates returned from Gemini")
	}

	// Extract token usage
	var tokenUsage *models.TokenUsage
	if resp.UsageMetadata != nil {
		tokenUsage = &models.TokenUsage{
			InputTokens:  resp.UsageMetadata.PromptTokenCount,
			OutputTokens: resp.UsageMetadata.CandidatesTokenCount,
			TotalTokens:  resp.UsageMetadata.TotalTokenCount,
		}
		log.Printf("Token usage - Input: %d, Output: %d, Total: %d", 
			tokenUsage.InputTokens, tokenUsage.OutputTokens, tokenUsage.TotalTokens)
	}

	// Extract text from response
	var responseText string
	for _, part := range resp.Candidates[0].Content.Parts {
		if txt, ok := part.(genai.Text); ok {
			responseText += string(txt)
		}
	}

	// Clean the response text
	responseText = strings.TrimSpace(responseText)
	
	// Check if response is empty
	if responseText == "" {
		log.Printf("Empty response from Gemini. Candidate count: %d", len(resp.Candidates))
		if len(resp.Candidates) > 0 {
			log.Printf("Candidate finish reason: %v", resp.Candidates[0].FinishReason)
			if resp.Candidates[0].SafetyRatings != nil {
				log.Printf("Safety ratings: %v", resp.Candidates[0].SafetyRatings)
			}
		}
		return nil, fmt.Errorf("empty response from Gemini AI")
	}
	
	// Remove markdown code blocks if present
	if strings.HasPrefix(responseText, "```json") {
		responseText = strings.TrimPrefix(responseText, "```json")
		responseText = strings.TrimSuffix(responseText, "```")
		responseText = strings.TrimSpace(responseText)
	} else if strings.HasPrefix(responseText, "```") {
		responseText = strings.TrimPrefix(responseText, "```")
		responseText = strings.TrimSuffix(responseText, "```")
		responseText = strings.TrimSpace(responseText)
	}

	log.Printf("Gemini response (%d chars): %s", len(responseText), responseText)

	// Parse JSON response
	var newsResponse models.NewsResponse
	if err := json.Unmarshal([]byte(responseText), &newsResponse); err != nil {
		return nil, fmt.Errorf("failed to parse Gemini response as JSON: %w\nResponse: %s", err, responseText)
	}

	// Validate response
	if len(newsResponse.News) == 0 {
		return nil, fmt.Errorf("Gemini returned no news items")
	}

	// Ensure we have at most 5 items for both types (to reduce output tokens)
	maxItems := 5
	if len(newsResponse.News) > maxItems {
		newsResponse.News = newsResponse.News[:maxItems]
	}

	// Add token usage to response
	newsResponse.TokenUsage = tokenUsage

	log.Printf("Gemini processed %d articles and returned %d top AI news items", len(newsItems), len(newsResponse.News))

	return &newsResponse, nil
}