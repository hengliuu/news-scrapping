# CURL Guide - News Scrapping Service API

Complete guide for testing the AI Tech News Scrapping Service using CURL commands.

## 🚀 Quick Start

First, make sure the service is running:
```bash
# Local development
go run main.go

# Or with Docker
docker-compose up -d
```

The service runs on `http://localhost:6005` by default.

## 📋 Available Endpoints

### 1. Health Check
**Endpoint**: `GET /health` or `GET /api/v1/health`  
**Purpose**: Check if the service is running

```bash
# Basic health check
curl http://localhost:6005/health

# API v1 health check
curl http://localhost:6005/api/v1/health

# With pretty JSON formatting
curl -s http://localhost:6005/health | jq

# Check response headers
curl -i http://localhost:6005/health
```

**Expected Response:**
```json
{
  "service": "news-scrapping-service",
  "status": "healthy",
  "timestamp": "2024-08-04T12:30:00Z",
  "version": "1.0.0"
}
```

### 2. Get Service Status
**Endpoint**: `GET /api/v1/status`  
**Purpose**: Get last job execution status and next scheduled run

```bash
# Get current status
curl http://localhost:6005/api/v1/status

# With pretty formatting
curl -s http://localhost:6005/api/v1/status | jq

# Save response to file
curl -s http://localhost:6005/api/v1/status | jq > status.json
```

**Expected Response:**
```json
{
  "last_run": "2024-08-04T08:00:00+07:00",
  "status": "success",
  "news_count": 5,
  "next_run": "2024-08-05T08:00:00+07:00",
  "error": ""
}
```

**Status Values:**
- `initialized` - Service just started
- `running` - Job currently executing
- `success` - Last job completed successfully
- `failed` - Last job failed (check error field)

### 3. Manual News Trigger
**Endpoint**: `POST /api/v1/trigger`  
**Purpose**: Manually trigger news scraping and Discord delivery

```bash
# AI Tech News (default)
curl -X POST http://localhost:6005/api/v1/trigger

# Global Tech/Business News
curl -X POST "http://localhost:6005/api/v1/trigger?type=global"

# With headers for better response
curl -X POST \
  -H "Content-Type: application/json" \
  "http://localhost:6005/api/v1/trigger?type=ai"

# With pretty output
curl -s -X POST "http://localhost:6005/api/v1/trigger?type=global" | jq

# Check if job is already running
curl -X POST http://localhost:6005/api/v1/trigger -w "\nHTTP Status: %{http_code}\n"
```

**Expected Response (Success):**
```json
{
  "message": "News scraping job triggered successfully",
  "data": {
    "triggered_at": "2024-08-04T12:30:00Z"
  }
}
```

**Expected Response (Already Running):**
```json
{
  "message": "News job is already running",
  "error": "Job in progress"
}
```
*HTTP Status: 429 (Too Many Requests)*

### 4. Get Latest News
**Endpoint**: `GET /api/v1/latest`  
**Purpose**: Get latest news without sending to Discord (for testing)

```bash
# AI Tech News (default)
curl http://localhost:6005/api/v1/latest

# Global Tech/Business News
curl "http://localhost:6005/api/v1/latest?type=global"

# With pretty formatting
curl -s "http://localhost:6005/api/v1/latest?type=ai" | jq

# Save to file for analysis
curl -s "http://localhost:6005/api/v1/latest?type=global" | jq > global_news.json

# Get only the news array
curl -s http://localhost:6005/api/v1/latest | jq '.data.news'

# Count articles by type
curl -s "http://localhost:6005/api/v1/latest?type=ai" | jq '.data.news | length'

# Filter global news by source
curl -s "http://localhost:6005/api/v1/latest?type=global" | jq '.data.news[] | select(.source | contains("Bloomberg"))'
```

**Expected Response:**
```json
{
  "message": "Latest AI tech news retrieved successfully",
  "data": {
    "news": [
      {
        "title": "OpenAI Announces GPT-5 Development",
        "summary": "OpenAI has officially announced the development of GPT-5...",
        "url": "https://techcrunch.com/example-article",
        "source": "TechCrunch AI",
        "relevance": "Major breakthrough in AI capabilities that will impact developers worldwide"
      },
      {
        "title": "Google's New AI Chip Breakthrough",
        "summary": "Google unveils next-generation AI processing chip...",
        "url": "https://theverge.com/example-article",
        "source": "The Verge AI",
        "relevance": "Hardware advancement that will accelerate AI model training and inference"
      }
    ],
    "generated_at": "2024-08-04T12:30:00Z",
    "scraped_count": 23,
    "selected_count": 5,
    "source_count": 5
  }
}
```

### 5. Root Endpoint
**Endpoint**: `GET /`  
**Purpose**: Service information and available endpoints

```bash
# Get service info
curl http://localhost:6005/

# Pretty format
curl -s http://localhost:6005/ | jq
```

## 🔄 Complete Testing Workflow

### Testing Sequence
Run these commands in order to fully test the service:

```bash
#!/bin/bash

echo "=== Testing News Scrapping Service ==="

# 1. Check if service is healthy
echo "1. Health Check:"
curl -s http://localhost:6005/health | jq
echo ""

# 2. Get current status
echo "2. Current Status:"
curl -s http://localhost:6005/api/v1/status | jq
echo ""

# 3. Test AI News
echo "3. AI Tech News (Testing Mode):"
curl -s http://localhost:6005/api/v1/latest | jq '.data | {type, news_count: (.news | length), scraped_count, source_count}'
echo ""

# 4. Test Global News
echo "4. Global Tech News (Testing Mode):"
curl -s "http://localhost:6005/api/v1/latest?type=global" | jq '.data | {type, news_count: (.news | length), scraped_count, source_count}'
echo ""

# 5. Trigger AI news job
echo "5. Trigger AI News:"
curl -s -X POST http://localhost:6005/api/v1/trigger | jq
echo ""

# 6. Wait and check status
echo "6. Waiting 5 seconds then checking status..."
sleep 5
curl -s http://localhost:6005/api/v1/status | jq
echo ""

# 7. Trigger Global news job
echo "7. Trigger Global News:"
curl -s -X POST "http://localhost:6005/api/v1/trigger?type=global" | jq
echo ""

echo "=== Testing Complete ==="
```

Save this as `test_api.sh` and run:
```bash
chmod +x test_api.sh
./test_api.sh
```

## 🐛 Troubleshooting with CURL

### Check if Service is Running
```bash
# Test connection
curl -f http://localhost:6005/health > /dev/null && echo "Service is UP" || echo "Service is DOWN"

# Check which port is in use
curl -s http://localhost:6005/health || curl -s http://localhost:8080/health || echo "Service not found"
```

### Debug Response Headers
```bash
# Get full HTTP response
curl -i http://localhost:6005/api/v1/status

# Check CORS headers
curl -i -X OPTIONS http://localhost:6005/api/v1/status

# Test with different User-Agent
curl -H "User-Agent: NewsBot/1.0" http://localhost:6005/health
```

### Monitor Job Execution
```bash
# Watch status changes during job execution
watch -n 2 'curl -s http://localhost:6005/api/v1/status | jq ".status, .last_run"'

# Monitor while triggering job
curl -X POST http://localhost:6005/api/v1/trigger && watch -n 1 'curl -s http://localhost:6005/api/v1/status | jq'
```

### Performance Testing
```bash
# Measure response time
curl -w "@curl-format.txt" -s http://localhost:6005/api/v1/latest

# Create curl-format.txt file:
cat > curl-format.txt << EOF
     time_namelookup:  %{time_namelookup}\n
        time_connect:  %{time_connect}\n
     time_appconnect:  %{time_appconnect}\n
    time_pretransfer:  %{time_pretransfer}\n
       time_redirect:  %{time_redirect}\n
  time_starttransfer:  %{time_starttransfer}\n
                     ----------\n
          time_total:  %{time_total}\n
EOF
```

## 📊 Response Status Codes

| Code | Meaning | When It Occurs |
|------|---------|----------------|
| 200 | OK | Successful request |
| 429 | Too Many Requests | Job already running (POST /trigger) |
| 500 | Internal Server Error | Service error (check logs) |
| 404 | Not Found | Invalid endpoint |

## 🔍 Advanced CURL Usage

### JSON Data Analysis
```bash
# Extract specific fields from news
curl -s http://localhost:6005/api/v1/latest | jq '.data.news[] | {title, source}'

# Get only titles
curl -s http://localhost:6005/api/v1/latest | jq -r '.data.news[].title'

# Count articles by source
curl -s http://localhost:6005/api/v1/latest | jq '.data.news | group_by(.source) | map({source: .[0].source, count: length})'

# Get token usage statistics
curl -s http://localhost:6005/api/v1/latest | jq '.data.token_usage'

# Monitor token usage over multiple requests
curl -s http://localhost:6005/api/v1/latest | jq '{type: .data.type, tokens: .data.token_usage.total_tokens}'
```

### Automated Monitoring
```bash
# Create monitoring script
cat > monitor.sh << 'EOF'
#!/bin/bash
while true; do
    STATUS=$(curl -s http://localhost:6005/api/v1/status | jq -r '.status')
    echo "$(date): Status = $STATUS"
    if [ "$STATUS" = "running" ]; then
        echo "Job is running..."
    fi
    sleep 10
done
EOF

chmod +x monitor.sh
./monitor.sh
```

### Error Handling
```bash
# Handle connection errors
curl --connect-timeout 5 --max-time 10 http://localhost:6005/health || echo "Connection failed"

# Retry on failure
curl --retry 3 --retry-delay 2 http://localhost:6005/api/v1/latest
```

## 📝 Example Use Cases

### Daily Operations Check
```bash
# Morning check before work
curl -s http://localhost:6005/api/v1/status | jq '{status, last_run, news_count, next_run}'
```

### Manual News Update
```bash
# Get fresh AI news immediately
curl -X POST http://localhost:6005/api/v1/trigger
sleep 10
curl -s http://localhost:6005/api/v1/status | jq

# Get fresh global news
curl -X POST "http://localhost:6005/api/v1/trigger?type=global"
```

### Development Testing
```bash
# Test AI news without Discord (safe for development)
curl -s http://localhost:6005/api/v1/latest | jq '.data.news[0]'

# Test global news without Discord
curl -s "http://localhost:6005/api/v1/latest?type=global" | jq '.data.news[0]'

# Compare both news types
echo "AI News Sources:"
curl -s http://localhost:6005/api/v1/latest | jq -r '.data.news[].source' | sort | uniq
echo "\nGlobal News Sources:"
curl -s "http://localhost:6005/api/v1/latest?type=global" | jq -r '.data.news[].source' | sort | uniq
```

## 🎯 Pro Tips

1. **Use `jq` for JSON parsing** - Install with `brew install jq` or `apt install jq`
2. **Save responses** - Use `curl -s url | jq > response.json` for analysis
3. **Monitor continuously** - Use `watch` command for real-time monitoring
4. **Check logs** - If API returns errors, check application logs
5. **Test locally first** - Always test `/api/v1/latest` before `/api/v1/trigger`

---

**Ready to test your News Scrapping Service! 🚀**