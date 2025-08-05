package models

import "time"

// NewsItem represents a single news article
type NewsItem struct {
	Title     string `json:"title"`
	Summary   string `json:"summary"`
	URL       string `json:"url"`
	Source    string `json:"source"`
	Relevance string `json:"relevance,omitempty"`
	PublishedAt time.Time `json:"published_at,omitempty"`
}

// NewsResponse represents the response from Gemini AI
type NewsResponse struct {
	News       []NewsItem  `json:"news"`
	TokenUsage *TokenUsage `json:"token_usage,omitempty"`
}

// TokenUsage represents token usage statistics from AI processing
type TokenUsage struct {
	InputTokens  int32 `json:"input_tokens"`
	OutputTokens int32 `json:"output_tokens"`
	TotalTokens  int32 `json:"total_tokens"`
}

// JobStatus represents the status of a news scraping job
type JobStatus struct {
	LastRun    time.Time `json:"last_run"`
	Status     string    `json:"status"`
	NewsCount  int       `json:"news_count"`
	NextRun    string    `json:"next_run"`
	Error      string    `json:"error,omitempty"`
}

// APIResponse represents a standard API response
type APIResponse struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}