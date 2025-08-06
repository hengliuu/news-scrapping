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
	client       *genai.Client
	model        *genai.GenerativeModel
	maxNewsItems int
}

// New creates a new Gemini AI client
func New(apiKey string) (*Client, error) {
	return NewWithConfig(apiKey, 5) // Default to 5 for backward compatibility
}

// NewWithConfig creates a new Gemini AI client with configurable max news items
func NewWithConfig(apiKey string, maxNewsItems int) (*Client, error) {
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
		client:       client,
		model:        model,
		maxNewsItems: maxNewsItems,
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

	// Limit news items to prevent overwhelming the AI and ensure quality processing
	var maxArticles int
	if newsType == "global" {
		maxArticles = 20 // Global news can handle more with default token limit
	} else {
		maxArticles = 15 // AI news - increased for better selection quality with top 10 output
	}

	if len(newsItems) > maxArticles {
		log.Printf("Limiting news items from %d to %d to prevent token overflow and ensure quality processing", len(newsItems), maxArticles)
		newsItems = newsItems[:maxArticles]
	}

	// Validate we have sufficient articles for meaningful curation
	minArticlesRequired := c.maxNewsItems + 2 // Need at least 2 more than output for meaningful selection
	if len(newsItems) < minArticlesRequired {
		log.Printf("Warning: Only %d articles available for selecting top %d. Consider adjusting news sources or filtering criteria", len(newsItems), c.maxNewsItems)
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

	// Create type-specific optimized prompts
	var prompt string
	if newsType == "global" {
		prompt = fmt.Sprintf(`You are an expert business and technology news curator for a daily Discord newsletter. Select the TOP %d most significant global business, technology, and cryptocurrency developments.

## EVALUATION CRITERIA (in order of priority):

1. **MARKET IMPACT** (40%% weight): Major market movements, IPOs, significant business decisions
2. **INNOVATION** (25%% weight): New tech products, crypto developments, breakthrough innovations  
3. **RECENCY** (20%% weight): Prefer articles from last 24-48 hours
4. **GLOBAL SIGNIFICANCE** (15%% weight): Stories affecting multiple markets or regions

## SELECTION RULES:
✅ INCLUDE: Reputable publications, official announcements, market-moving news
❌ EXCLUDE: Duplicates, opinion pieces, unverified rumors, articles >7 days old

Return EXACTLY this JSON with %d items ranked by importance:

%s

{"news":[{"title":"Clear headline (max 100 chars)","summary":"Key facts and implications (max 250 chars)","url":"original_url","source":"publication","relevance":"Why significant (max 100 chars)"}]}`, c.maxNewsItems, string(articlesJSON))
	} else {
		// AI-focused optimized prompt
		prompt = fmt.Sprintf(`You are an expert AI technology news curator for a daily Discord newsletter. Your task is to analyze the provided news articles and select the TOP %d most significant AI technology developments.

## EVALUATION CRITERIA (in order of priority):

1. **IMPACT SIGNIFICANCE** (40%% weight)
   - Major product launches or updates from leading AI companies
   - Breakthrough research publications or discoveries
   - Significant funding rounds or acquisitions in AI
   - New AI regulations or policy changes
   - Industry partnerships or collaborations

2. **RECENCY & RELEVANCE** (25%% weight)
   - Prefer articles published within the last 24-48 hours
   - Breaking news takes priority over older stories
   - Ongoing developments with new updates

3. **TECHNICAL INNOVATION** (20%% weight)
   - New AI model architectures or capabilities
   - Novel applications of existing AI technology
   - Performance benchmarks or comparisons
   - Open-source releases or tools

4. **BUSINESS & MARKET IMPACT** (15%% weight)
   - Market-moving announcements
   - Strategic business decisions
   - Industry adoption trends

## SELECTION RULES:
✅ INCLUDE: Reputable tech publications, official company announcements, major AI model updates, regulatory developments
❌ EXCLUDE: Duplicate stories, opinion pieces without new info, marketing content, unverified rumors, articles >7 days old

## DUPLICATE HANDLING:
If multiple articles cover the same story, select the most comprehensive and recent version from official sources.

Return EXACTLY this JSON structure with %d items ranked by importance:

%s

{"news":[{"title":"Clear, engaging headline (max 100 chars)","summary":"Concise 2-3 sentence summary focusing on key facts and implications (max 250 chars)","url":"original_article_url","source":"publication_name","relevance":"Brief explanation of why this is significant (max 100 chars)"}]}`, c.maxNewsItems, string(articlesJSON))
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
	if after, found := strings.CutPrefix(responseText, "```json"); found {
		responseText = strings.TrimSuffix(after, "```")
		responseText = strings.TrimSpace(responseText)
	} else if after, found := strings.CutPrefix(responseText, "```"); found {
		responseText = strings.TrimSuffix(after, "```")
		responseText = strings.TrimSpace(responseText)
	}

	log.Printf("Gemini response (%d chars): %s", len(responseText), responseText)

	// Parse JSON response with improved error handling
	var newsResponse models.NewsResponse
	if err := json.Unmarshal([]byte(responseText), &newsResponse); err != nil {
		// Try to extract JSON from potentially malformed response
		if jsonStart := strings.Index(responseText, "{"); jsonStart >= 0 {
			if jsonEnd := strings.LastIndex(responseText, "}"); jsonEnd > jsonStart {
				cleanJSON := responseText[jsonStart : jsonEnd+1]
				log.Printf("Attempting to parse extracted JSON: %s", cleanJSON)
				if retryErr := json.Unmarshal([]byte(cleanJSON), &newsResponse); retryErr == nil {
					log.Printf("Successfully parsed extracted JSON")
				} else {
					return nil, fmt.Errorf("failed to parse Gemini response as JSON (retry also failed): %w\nOriginal response: %s", err, responseText)
				}
			} else {
				return nil, fmt.Errorf("failed to parse Gemini response as JSON: %w\nResponse: %s", err, responseText)
			}
		} else {
			return nil, fmt.Errorf("failed to parse Gemini response as JSON: %w\nResponse: %s", err, responseText)
		}
	}

	// Validate response
	if len(newsResponse.News) == 0 {
		return nil, fmt.Errorf("Gemini returned no news items. This may indicate low-quality input articles or overly restrictive filtering")
	}

	// Ensure we have at most the configured number of items
	if len(newsResponse.News) > c.maxNewsItems {
		newsResponse.News = newsResponse.News[:c.maxNewsItems]
	}

	// Add token usage to response
	newsResponse.TokenUsage = tokenUsage

	log.Printf("Gemini processed %d articles and returned %d top news items", len(newsItems), len(newsResponse.News))

	return &newsResponse, nil
}
