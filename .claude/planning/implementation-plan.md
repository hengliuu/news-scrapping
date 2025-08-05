# News Scrapping Service - Implementation Plan

## Project Overview
Daily AI Tech News delivered to Discord using Go, Gin Gonic, and Gemini AI (gemini-2.5-flash).

## Architecture Overview

```
news-scrapping/
├── main.go                 # Application entry point
├── go.mod                  # Go module file
├── go.sum                  # Go dependencies
├── .env                    # Environment variables
├── .gitignore             # Git ignore file
├── Dockerfile             # Docker configuration
├── README.md              # Project documentation
├── internal/              # Private application code
│   ├── config/            # Configuration management
│   │   └── config.go      # Config loader
│   ├── scraper/           # News scraping logic
│   │   ├── scraper.go     # Main scraper interface
│   │   └── sources.go     # News sources definitions
│   ├── ai/                # Gemini AI integration
│   │   ├── client.go      # Gemini API client
│   │   └── processor.go   # Content processing
│   ├── discord/           # Discord integration
│   │   └── webhook.go     # Discord webhook client
│   ├── scheduler/         # Cron job management
│   │   └── cron.go        # Scheduler implementation
│   └── api/               # REST API handlers
│       ├── router.go      # API routes
│       ├── handlers.go    # Request handlers
│       └── middleware.go  # API middleware
└── pkg/                   # Public packages
    └── models/            # Data models
        └── news.go        # News item structures
```

## Implementation Checklist

### Phase 1: Project Setup ✅
- [x] Initialize Go module: `go mod init github.com/hengky/news-scrapping`
- [x] Create basic project structure
- [x] Set up Git repository
- [x] Create .gitignore file
- [x] Create .env.example file

### Phase 2: Core Components ✅
- [x] **Configuration Module**
  - [x] Create config struct with all settings
  - [x] Load environment variables
  - [x] Validate required configurations
  
- [x] **Gin Web Server**
  - [x] Initialize Gin router
  - [x] Set up server on port 6005
  - [x] Add basic middleware (CORS, logging)
  - [x] Create graceful shutdown

### Phase 3: News Scraping ✅
- [x] **News Sources Setup**
  - [x] Define AI tech news sources (TechCrunch AI, The Verge AI, AI News, etc.)
  - [x] Create source interface
  - [x] Implement RSS/Web scraping for each source
  
- [x] **Content Extraction**
  - [x] Parse article titles
  - [x] Extract article summaries
  - [x] Get article URLs
  - [x] Filter for AI-related content

### Phase 4: Gemini AI Integration ✅
- [x] **Gemini Client**
  - [x] Set up Google AI SDK
  - [x] Create Gemini client with API key
  - [x] Handle rate limiting
  
- [x] **Content Processing**
  - [x] Create prompt template for Top 5 selection
  - [x] Process scraped news through Gemini
  - [x] Format AI response for Discord

### Phase 5: Discord Integration ✅
- [x] **Webhook Client**
  - [x] Create Discord webhook sender
  - [x] Format messages with embeds
  - [x] Add error handling for failed sends
  
- [x] **Message Formatting**
  - [x] Create rich embed format
  - [x] Include article titles, summaries, and links
  - [x] Add timestamp and branding

### Phase 6: Scheduling ✅
- [x] **Cron Setup**
  - [x] Add robfig/cron dependency
  - [x] Configure timezone to WIB (Asia/Jakarta)
  - [x] Schedule daily run at 08:00 WIB
  
- [x] **Job Management**
  - [x] Create news fetch job
  - [x] Add job status tracking
  - [x] Implement retry logic

### Phase 7: REST API ✅
- [x] **Endpoints**
  - [x] GET /health - Health check
  - [x] POST /trigger - Manual news fetch
  - [x] GET /status - Last run status
  - [x] GET /latest - Get latest news without sending
  
- [x] **API Security**
  - [x] Add CORS middleware
  - [x] Request validation
  - [x] Error handling

### Phase 8: Testing & Deployment ✅
- [x] **Testing**
  - [x] Build verification successful
  - [x] All dependencies resolved
  - [x] Code compilation verified
  
- [x] **Docker Setup**
  - [x] Create multi-stage Dockerfile
  - [x] Add docker-compose.yml
  - [x] Configure environment variables
  
- [x] **Documentation**
  - [x] Comprehensive API documentation
  - [x] Setup instructions
  - [x] Architecture overview
  - [x] Troubleshooting guide

## API Design

### Endpoints

```yaml
GET /health
  Response: 
    - 200: { "status": "healthy", "version": "1.0.0" }

POST /trigger
  Headers:
    - X-API-Key: {optional_api_key}
  Response:
    - 200: { "message": "News fetch triggered", "job_id": "uuid" }
    - 429: { "error": "Job already running" }

GET /status
  Response:
    - 200: { 
        "last_run": "2024-01-10T08:00:00+07:00",
        "status": "success",
        "news_count": 5,
        "next_run": "2024-01-11T08:00:00+07:00"
      }

GET /latest
  Response:
    - 200: {
        "news": [
          {
            "title": "OpenAI Announces GPT-5",
            "summary": "...",
            "url": "https://...",
            "source": "TechCrunch"
          }
        ],
        "generated_at": "2024-01-10T08:00:00+07:00"
      }
```

## News Processing Flow

```mermaid
graph TD
    A[Cron Trigger 08:00 WIB] --> B[Fetch from News Sources]
    B --> C[Extract AI Tech Articles]
    C --> D[Send to Gemini AI]
    D --> E[Get Top 5 News]
    E --> F[Format for Discord]
    F --> G[Send to Discord Webhook]
    G --> H[Update Status]
```

## Gemini Prompt Template

```
You are an AI news curator. From the following tech news articles, select the TOP 5 most important AI-related news for today. Focus on:
- Breakthrough AI technologies
- Major AI company announcements
- AI research developments
- AI industry trends
- AI applications and implementations

Articles:
{articles_json}

Return the top 5 in this JSON format:
{
  "news": [
    {
      "title": "Article title",
      "summary": "Brief 2-3 sentence summary",
      "url": "article URL",
      "relevance": "Why this is important for AI community"
    }
  ]
}
```

## Discord Message Format

```markdown
🤖 **Daily AI Tech News** - {date}

**1. {title}**
{summary}
🔗 [Read more]({url})

**2. {title}**
{summary}
🔗 [Read more]({url})

...

_Generated by AI News Bot powered by Gemini_
```

## Environment Configuration

```env
# API Keys
GEMINI_API_KEY=AIzaSyBCQOPdmnjmiiRGGjciFEkqLRVDC_N1cV8
DISCORD_WEBHOOK=https://discord.com/api/webhooks/...

# Server Configuration
PORT=6005
GIN_MODE=release

# Timezone
TZ=Asia/Jakarta

# Optional
API_KEY=your_api_key_for_manual_triggers
LOG_LEVEL=info
```

## Dependencies

```go
// go.mod
module github.com/hengky/news-scrapping

go 1.21

require (
    github.com/gin-gonic/gin v1.9.1
    github.com/google/generative-ai-go v0.11.0
    github.com/robfig/cron/v3 v3.0.1
    github.com/joho/godotenv v1.5.1
    github.com/mmcdole/gofeed v1.2.1  // For RSS parsing
)
```

## Error Handling Strategy

1. **Scraping Failures**: Log and continue with available sources
2. **AI API Failures**: Implement exponential backoff retry
3. **Discord Failures**: Queue for retry, alert via logs
4. **Cron Failures**: Log extensively, implement health checks

## Monitoring & Logging

- Structured logging with levels (info, warn, error)
- Log rotation for production
- Metrics for:
  - Daily job success/failure
  - API response times
  - News source availability
  - Gemini API usage

## Security Considerations

- Environment variables for sensitive data
- No hardcoded credentials
- Validate all external inputs
- Rate limiting on API endpoints
- Secure Discord webhook URL

## Future Enhancements

1. Multiple Discord channels support
2. User preference for news categories
3. Historical news archive
4. News source reliability scoring
5. Custom scheduling per channel
6. Web dashboard for monitoring