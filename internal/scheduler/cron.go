package scheduler

import (
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/hengky/news-scrapping/internal/ai"
	"github.com/hengky/news-scrapping/internal/config"
	"github.com/hengky/news-scrapping/internal/discord"
	"github.com/hengky/news-scrapping/internal/scraper"
	"github.com/hengky/news-scrapping/pkg/models"
	"github.com/robfig/cron/v3"
)

// Scheduler handles scheduled tasks
type Scheduler struct {
	cron          *cron.Cron
	config        *config.Config
	scraper       *scraper.Scraper
	aiProcessor   *ai.Processor
	discord       *discord.WebhookClient
	discordGlobal *discord.WebhookClient
	jobStatus     *models.JobStatus
	mu            sync.RWMutex
	running       bool
}

// New creates a new scheduler
func New(cfg *config.Config) *Scheduler {
	// Create timezone location
	location, err := time.LoadLocation(cfg.Timezone)
	if err != nil {
		log.Printf("Warning: Could not load timezone %s, using UTC: %v", cfg.Timezone, err)
		location = time.UTC
	}

	// Create cron with timezone
	c := cron.New(cron.WithLocation(location))

	// Initialize components
	scraperInstance := scraper.New()

	aiProcessor, err := ai.NewProcessor(cfg)
	if err != nil {
		log.Fatalf("Failed to create AI processor: %v", err)
	}

	discordClient := discord.New(cfg.DiscordWebhook)
	discordGlobalClient := discord.New(cfg.DiscordWebhookGlobal)

	return &Scheduler{
		cron:          c,
		config:        cfg,
		scraper:       scraperInstance,
		aiProcessor:   aiProcessor,
		discord:       discordClient,
		discordGlobal: discordGlobalClient,
		jobStatus: &models.JobStatus{
			Status:    "initialized",
			NewsCount: 0,
			NextRun:   "08:00 WIB daily",
		},
	}
}

// Start starts the scheduler
func (s *Scheduler) Start() {
	// Schedule daily job at 08:00 WIB
	_, err := s.cron.AddFunc("0 8 * * *", s.runNewsJob)
	if err != nil {
		log.Fatalf("Failed to schedule news job: %v", err)
	}

	s.cron.Start()
	log.Printf("Scheduler started - News job scheduled for 08:00 %s daily", s.config.Timezone)

	// Update next run time
	s.updateNextRunTime()
}

// Stop stops the scheduler
func (s *Scheduler) Stop() {
	if s.cron != nil {
		s.cron.Stop()
	}
	if s.aiProcessor != nil {
		s.aiProcessor.Close()
	}
	log.Println("Scheduler stopped")
}

// RunManualJob runs the AI news job manually (backward compatibility)
func (s *Scheduler) RunManualJob() error {
	return s.RunManualJobByType("ai")
}

// RunManualJobByType runs the news job manually with specified type
func (s *Scheduler) RunManualJobByType(newsType string) error {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return fmt.Errorf("job is already running")
	}
	s.running = true
	s.mu.Unlock()

	defer func() {
		s.mu.Lock()
		s.running = false
		s.mu.Unlock()
	}()

	return s.executeNewsJobByType(newsType)
}

// runNewsJob is the scheduled job function
func (s *Scheduler) runNewsJob() {
	s.mu.Lock()
	if s.running {
		log.Println("News job already running, skipping this execution")
		s.mu.Unlock()
		return
	}
	s.running = true
	s.mu.Unlock()

	defer func() {
		s.mu.Lock()
		s.running = false
		s.mu.Unlock()
		s.updateNextRunTime()
	}()

	log.Println("Starting scheduled news job...")

	if err := s.executeNewsJob(); err != nil {
		log.Printf("Scheduled news job failed: %v", err)

		// Send error notification to Discord
		errorMsg := fmt.Sprintf("❌ **News Bot Error**\n\nScheduled job failed at %s\n\nError: %s",
			time.Now().Format("2006-01-02 15:04:05 MST"), err.Error())

		if discordErr := s.discord.SendSimpleMessage(errorMsg); discordErr != nil {
			log.Printf("Failed to send error notification to Discord: %v", discordErr)
		}
	} else {
		log.Println("Scheduled news job completed successfully")
	}
}

// executeNewsJob executes the complete AI news processing pipeline (backward compatibility)
func (s *Scheduler) executeNewsJob() error {
	return s.executeNewsJobByType("ai")
}

// executeNewsJobByType executes the complete news processing pipeline for a specific type
func (s *Scheduler) executeNewsJobByType(newsType string) error {
	startTime := time.Now()

	s.mu.Lock()
	s.jobStatus.Status = "running"
	s.jobStatus.Error = ""
	s.mu.Unlock()

	// Step 1: Scrape news from sources based on type
	log.Printf("Step 1: Scraping %s news from sources...", newsType)
	newsItems, err := s.scraper.ScrapeNewsByType(newsType)
	if err != nil {
		s.updateJobStatus("failed", 0, err.Error())
		return fmt.Errorf("failed to scrape %s news: %w", newsType, err)
	}

	if len(newsItems) == 0 {
		s.updateJobStatus("completed", 0, fmt.Sprintf("No %s news items found", newsType))
		return fmt.Errorf("no %s news items scraped", newsType)
	}

	log.Printf("Scraped %d %s news items", len(newsItems), newsType)

	// Step 2: Process with AI to get top 5
	log.Printf("Step 2: Processing %s news with Gemini AI...", newsType)
	newsResponse, err := s.aiProcessor.ProcessNewsItemsByType(newsItems, newsType)
	if err != nil {
		s.updateJobStatus("failed", 0, err.Error())
		return fmt.Errorf("failed to process %s news with AI: %w", newsType, err)
	}

	if len(newsResponse.News) == 0 {
		s.updateJobStatus("completed", 0, fmt.Sprintf("No relevant %s news found", newsType))
		return fmt.Errorf("AI processing returned no %s news items", newsType)
	}

	log.Printf("AI selected %d top %s news items", len(newsResponse.News), newsType)

	// Step 3: Send to Discord (use appropriate webhook)
	log.Printf("Step 3: Sending %s news to Discord...", newsType)
	var discordErr error
	if newsType == "global" {
		log.Printf("Using global Discord webhook for %s news", newsType)
		discordErr = s.discordGlobal.SendNewsByType(newsResponse, newsType)
	} else {
		log.Printf("Using AI Discord webhook for %s news", newsType)
		discordErr = s.discord.SendNewsByType(newsResponse, newsType)
	}

	if discordErr != nil {
		s.updateJobStatus("failed", len(newsResponse.News), discordErr.Error())
		return fmt.Errorf("failed to send %s news to Discord: %w", newsType, discordErr)
	}

	// Update job status
	s.updateJobStatus("success", len(newsResponse.News), "")

	duration := time.Since(startTime)
	log.Printf("%s news job completed successfully in %v - sent %d news items to Discord",
		strings.Title(newsType), duration, len(newsResponse.News))

	return nil
}

// GetJobStatus returns the current job status
func (s *Scheduler) GetJobStatus() *models.JobStatus {
	s.mu.RLock()
	defer s.mu.RUnlock()

	status := *s.jobStatus // Copy the status
	return &status
}

// IsRunning returns whether a job is currently running
func (s *Scheduler) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.running
}

// updateJobStatus updates the job status
func (s *Scheduler) updateJobStatus(status string, newsCount int, errorMsg string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.jobStatus.LastRun = time.Now()
	s.jobStatus.Status = status
	s.jobStatus.NewsCount = newsCount
	s.jobStatus.Error = errorMsg
}

// updateNextRunTime updates the next run time
func (s *Scheduler) updateNextRunTime() {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Calculate next 08:00 in the configured timezone
	location, err := time.LoadLocation(s.config.Timezone)
	if err != nil {
		location = time.UTC
	}

	now := time.Now().In(location)
	next := time.Date(now.Year(), now.Month(), now.Day(), 8, 0, 0, 0, location)

	// If it's already past 08:00 today, schedule for tomorrow
	if now.Hour() >= 8 {
		next = next.Add(24 * time.Hour)
	}

	s.jobStatus.NextRun = next.Format("2006-01-02 15:04:05 MST")
}
