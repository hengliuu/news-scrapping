# AI Tech News Scrapping Service

Daily AI Tech News delivered to Discord using Go, Gin Gonic, and Gemini AI.

## Features

- **Automated News Scraping**: Collects AI tech news from multiple sources (TechCrunch, The Verge, AI News, VentureBeat, MIT Technology Review)
- **Global News Support**: Also scrapes global tech/business news from Bloomberg, Reuters, Financial Times, and more
- **AI-Powered Curation**: Uses Gemini 2.5 Flash to select the top 5 most important news daily
- **Multi-Type Support**: Supports both AI-specific news and global tech/business news
- **Discord Integration**: Sends formatted news updates to Discord via webhook with type-specific styling
- **Scheduled Execution**: Runs daily at 08:00 WIB (Western Indonesia Time)
- **REST API**: Provides endpoints for manual triggers and status monitoring with news type parameters
- **Dockerized**: Ready for containerized deployment

## Architecture

```
news-scrapping/
├── main.go                 # Application entry point
├── internal/              # Private application code
│   ├── config/            # Configuration management
│   ├── scraper/           # News scraping logic
│   ├── ai/                # Gemini AI integration
│   ├── discord/           # Discord webhook client
│   ├── scheduler/         # Cron job management
│   └── api/               # REST API handlers
└── pkg/                   # Public packages
    └── models/            # Data models
```

## Quick Start

### Prerequisites

- Go 1.21+
- Gemini API Key
- Discord Webhook URL

### Environment Setup

1. Copy the environment template:
   ```bash
   cp .env.example .env
   ```

2. Configure your environment variables in `.env`:
   ```env
   GEMINI_API_KEY=your_gemini_api_key
   DISCORD_WEBHOOK=your_discord_webhook_url
   PORT=6005
   TZ=Asia/Jakarta
   ```

### Running Locally

1. Install dependencies:
   ```bash
   go mod download
   ```

2. Run the application:
   ```bash
   go run main.go
   ```

3. The service will start on port 6005 and schedule daily news updates at 08:00 WIB.

### Docker Deployment

1. Build and run with Docker Compose:
   ```bash
   docker-compose up -d
   ```

2. Or build manually:
   ```bash
   docker build -t news-scrapping .
   docker run -p 6005:6005 --env-file .env news-scrapping
   ```

## API Endpoints

### Health Check
```
GET /health
GET /api/v1/health
```
Returns service health status.

### Get Status
```
GET /api/v1/status
```
Returns the last job execution status and next scheduled run.

**Response:**
```json
{
  "last_run": "2024-01-10T08:00:00+07:00",
  "status": "success",
  "news_count": 5,
  "next_run": "2024-01-11T08:00:00+07:00",
  "error": ""
}
```

### Manual Trigger
```
POST /api/v1/trigger
POST /api/v1/trigger?type=ai     # AI tech news (default)
POST /api/v1/trigger?type=global # Global tech/business news
```
Manually triggers the news scraping and processing job.

**Query Parameters:**
- `type` (optional): News type to fetch - `ai` (default) or `global`

**Response:**
```json
{
  "message": "News scraping job triggered successfully",
  "data": {
    "triggered_at": "2024-01-10T10:30:00Z",
    "type": "ai"
  }
}
```

### Get Latest News
```
GET /api/v1/latest
GET /api/v1/latest?type=ai     # AI tech news (default)
GET /api/v1/latest?type=global # Global tech/business news
```
Retrieves the latest news without sending to Discord.

**Query Parameters:**
- `type` (optional): News type to fetch - `ai` (default) or `global`

**Response:**
```json
{
  "message": "Latest news retrieved successfully",
  "data": {
    "news": [
      {
        "title": "OpenAI Announces GPT-5",
        "summary": "OpenAI has announced the development of GPT-5...",
        "url": "https://example.com/news",
        "source": "TechCrunch AI",
        "relevance": "Major breakthrough in AI capabilities",
        "type": "ai"
      }
    ],
    "generated_at": "2024-01-10T10:30:00Z",
    "scraped_count": 25,
    "selected_count": 5,
    "source_count": 5,
    "type": "ai",
    "token_usage": {
      "input_tokens": 1234,
      "output_tokens": 456,
      "total_tokens": 1690
    }
  }
}
```

## Configuration

### Environment Variables

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `GEMINI_API_KEY` | Google Gemini API key | - | ✅ |
| `DISCORD_WEBHOOK` | Discord webhook URL | - | ✅ |
| `PORT` | Server port | 6005 | ❌ |
| `GIN_MODE` | Gin framework mode | release | ❌ |
| `TZ` | Timezone for scheduling | Asia/Jakarta | ❌ |
| `LOG_LEVEL` | Logging level | info | ❌ |

### News Sources

The service scrapes from different sources based on the news type:

#### AI Tech News Sources (`type=ai`):
- **TechCrunch AI**: Latest AI news from TechCrunch
- **The Verge AI**: AI coverage from The Verge
- **AI News**: Dedicated AI news source
- **VentureBeat AI**: AI business and technology news
- **MIT Technology Review AI**: Research and analysis

#### Global Tech/Business News Sources (`type=global`):
- **Bloomberg Technology**: Tech and business news from Bloomberg
- **Reuters Technology**: Global tech coverage
- **Financial Times Tech**: Technology and innovation news
- **Wall Street Journal Tech**: Business-focused tech news
- **CNBC Technology**: Market and tech news
- **BBC Technology**: Global technology coverage
- **The Guardian Tech**: UK and international tech news
- **Forbes Tech**: Business and technology insights

## Discord Message Format

The bot sends rich embedded messages to Discord with:

- **Daily header** with date and news type
- **Top 5 news items** as individual embeds with:
  - Article title and ranking
  - Brief summary
  - Relevance explanation (context-aware based on type)
  - Source attribution
  - Direct link to article
- **Visual distinction**:
  - **AI News**: Green color scheme (0x00D4AA)
  - **Global News**: Blue color scheme (0x1E88E5)
- **Bot signature** with Gemini AI attribution
- **Token usage statistics**: Shows input, output, and total tokens used

## Monitoring and Logging

### Health Checks

- Service health: `GET /health`
- Docker health check: Built-in container health monitoring
- Cron job status: Available via `/api/v1/status`

### Logging

The service provides structured logging for:

- News scraping operations
- AI processing results
- Discord webhook delivery
- API requests and responses
- Scheduled job execution
- Error tracking and recovery

## Error Handling

### Resilient Design

- **Source failures**: Continues with available sources if some fail
- **AI processing**: Implements retry logic with exponential backoff
- **Discord delivery**: Queues failed messages for retry
- **Graceful shutdown**: Properly closes connections and saves state

### Error Notifications

Failed scheduled jobs trigger error notifications sent to the Discord channel with:
- Timestamp of failure
- Error description
- Guidance for manual intervention

## Development

### Project Structure

```
internal/
├── config/         # Configuration loading and validation
├── scraper/        # Web scraping and RSS feed parsing
├── ai/            # Gemini AI client and processing
├── discord/       # Discord webhook integration
├── scheduler/     # Cron job management
└── api/           # HTTP handlers and routing

pkg/
└── models/        # Shared data structures
```

### Adding News Sources

To add new news sources:

1. Add source to `internal/scraper/sources.go`:
   ```go
   {
       Name: "New Source",
       URL:  "https://example.com/feed.xml",
       Type: "rss",
   }
   ```

2. Test with manual trigger: `POST /api/v1/trigger`

### Testing

Run tests with:
```bash
go test ./...
```

## Deployment

### Production Checklist

- [ ] Configure proper logging levels
- [ ] Set up monitoring and alerting
- [ ] Configure firewall rules for port 6005
- [ ] Set up SSL/TLS termination (if needed)
- [ ] Configure container restart policies
- [ ] Set up log rotation
- [ ] Test Discord webhook connectivity
- [ ] Verify Gemini API key permissions
- [ ] Test timezone configuration

### Scaling Considerations

- The service is designed for single-instance deployment
- Database is not required (stateless design)
- Can be deployed behind a load balancer for high availability
- Consider rate limits for Gemini API and Discord webhooks

## Troubleshooting

### Common Issues

1. **"No news items scraped"**
   - Check internet connectivity
   - Verify RSS feed URLs are accessible
   - Review source filtering criteria

2. **"Failed to process with AI"**
   - Verify Gemini API key is valid and active
   - Check API quota and billing status
   - Review API response in logs

3. **"Discord webhook failed"**
   - Verify webhook URL is correct and active
   - Check Discord server permissions
   - Test webhook with `/api/v1/test-discord` (if implemented)

4. **"Cron job not running"**
   - Verify timezone configuration
   - Check server system time
   - Review scheduler logs for errors

### Logs Location

- **Docker**: `docker logs <container_name>`
- **Local**: Console output with structured JSON logging

## API Examples

### Check Service Status
```bash
curl http://localhost:6005/api/v1/status
```

### Trigger Manual News Update
```bash
# AI tech news (default)
curl -X POST http://localhost:6005/api/v1/trigger

# Global tech/business news
curl -X POST "http://localhost:6005/api/v1/trigger?type=global"
```

### Get Latest News
```bash
# AI tech news (default)
curl http://localhost:6005/api/v1/latest

# Global tech/business news
curl "http://localhost:6005/api/v1/latest?type=global"
```

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## Support

For issues and questions:
- Check the troubleshooting section
- Review application logs
- Open an issue on GitHub

---

**Built with ❤️ using Go, Gin, and Gemini AI**