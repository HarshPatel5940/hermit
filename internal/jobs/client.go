package jobs

import (
	"context"
	"fmt"
	"time"

	"github.com/hibiken/asynq"
	"go.uber.org/zap"
)

// Client wraps asynq.Client for enqueuing tasks.
type Client struct {
	client *asynq.Client
	logger *zap.Logger
}

// NewClient creates a new job client.
func NewClient(redisURL string, logger *zap.Logger) (*Client, error) {
	opt, err := asynq.ParseRedisURI(redisURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse redis URL: %w", err)
	}

	client := asynq.NewClient(opt)

	logger.Info("Job client initialized", zap.String("redisURL", redisURL))

	return &Client{
		client: client,
		logger: logger,
	}, nil
}

// Close closes the job client.
func (c *Client) Close() error {
	return c.client.Close()
}

// EnqueueCrawlWebsite enqueues a crawl website task.
func (c *Client) EnqueueCrawlWebsite(ctx context.Context, websiteID uint, startURL string) error {
	payload, err := NewCrawlWebsitePayload(websiteID, startURL)
	if err != nil {
		return fmt.Errorf("failed to create crawl payload: %w", err)
	}

	task := asynq.NewTask(TypeCrawlWebsite, payload)

	info, err := c.client.EnqueueContext(ctx, task,
		asynq.MaxRetry(3),
		asynq.Timeout(30*time.Minute),
		asynq.Queue("crawl"),
	)
	if err != nil {
		c.logger.Error("Failed to enqueue crawl task",
			zap.Uint("websiteID", websiteID),
			zap.String("url", startURL),
			zap.Error(err),
		)
		return fmt.Errorf("failed to enqueue crawl task: %w", err)
	}

	c.logger.Info("Enqueued crawl task",
		zap.Uint("websiteID", websiteID),
		zap.String("url", startURL),
		zap.String("taskID", info.ID),
		zap.String("queue", info.Queue),
	)

	return nil
}

// EnqueueVectorizePage enqueues a vectorize page task.
func (c *Client) EnqueueVectorizePage(ctx context.Context, websiteID, pageID uint, pageURL, content string) error {
	payload, err := NewVectorizePagePayload(websiteID, pageID, pageURL, content)
	if err != nil {
		return fmt.Errorf("failed to create vectorize payload: %w", err)
	}

	task := asynq.NewTask(TypeVectorizePage, payload)

	info, err := c.client.EnqueueContext(ctx, task,
		asynq.MaxRetry(5),
		asynq.Timeout(10*time.Minute),
		asynq.Queue("vectorize"),
	)
	if err != nil {
		c.logger.Error("Failed to enqueue vectorize task",
			zap.Uint("websiteID", websiteID),
			zap.Uint("pageID", pageID),
			zap.String("url", pageURL),
			zap.Error(err),
		)
		return fmt.Errorf("failed to enqueue vectorize task: %w", err)
	}

	c.logger.Debug("Enqueued vectorize task",
		zap.Uint("websiteID", websiteID),
		zap.Uint("pageID", pageID),
		zap.String("taskID", info.ID),
	)

	return nil
}

// EnqueueRecrawlWebsite enqueues a recrawl website task.
func (c *Client) EnqueueRecrawlWebsite(ctx context.Context, websiteID uint) error {
	payload, err := NewRecrawlWebsitePayload(websiteID)
	if err != nil {
		return fmt.Errorf("failed to create recrawl payload: %w", err)
	}

	task := asynq.NewTask(TypeRecrawlWebsite, payload)

	info, err := c.client.EnqueueContext(ctx, task,
		asynq.MaxRetry(3),
		asynq.Timeout(30*time.Minute),
		asynq.Queue("crawl"),
	)
	if err != nil {
		c.logger.Error("Failed to enqueue recrawl task",
			zap.Uint("websiteID", websiteID),
			zap.Error(err),
		)
		return fmt.Errorf("failed to enqueue recrawl task: %w", err)
	}

	c.logger.Info("Enqueued recrawl task",
		zap.Uint("websiteID", websiteID),
		zap.String("taskID", info.ID),
	)

	return nil
}

// EnqueueCleanupOldPages enqueues a cleanup old pages task.
func (c *Client) EnqueueCleanupOldPages(ctx context.Context, websiteID uint, daysOld int, deleteFrom string) error {
	payload, err := NewCleanupOldPagesPayload(websiteID, daysOld, deleteFrom)
	if err != nil {
		return fmt.Errorf("failed to create cleanup payload: %w", err)
	}

	task := asynq.NewTask(TypeCleanupOldPages, payload)

	info, err := c.client.EnqueueContext(ctx, task,
		asynq.MaxRetry(2),
		asynq.Timeout(20*time.Minute),
		asynq.Queue("maintenance"),
	)
	if err != nil {
		c.logger.Error("Failed to enqueue cleanup task",
			zap.Uint("websiteID", websiteID),
			zap.Int("daysOld", daysOld),
			zap.Error(err),
		)
		return fmt.Errorf("failed to enqueue cleanup task: %w", err)
	}

	c.logger.Info("Enqueued cleanup task",
		zap.Uint("websiteID", websiteID),
		zap.Int("daysOld", daysOld),
		zap.String("taskID", info.ID),
	)

	return nil
}

// EnqueueCrawlWebsiteDelayed enqueues a crawl task with a delay.
func (c *Client) EnqueueCrawlWebsiteDelayed(ctx context.Context, websiteID uint, startURL string, delay time.Duration) error {
	payload, err := NewCrawlWebsitePayload(websiteID, startURL)
	if err != nil {
		return fmt.Errorf("failed to create crawl payload: %w", err)
	}

	task := asynq.NewTask(TypeCrawlWebsite, payload)

	info, err := c.client.EnqueueContext(ctx, task,
		asynq.MaxRetry(3),
		asynq.Timeout(30*time.Minute),
		asynq.Queue("crawl"),
		asynq.ProcessIn(delay),
	)
	if err != nil {
		c.logger.Error("Failed to enqueue delayed crawl task",
			zap.Uint("websiteID", websiteID),
			zap.String("url", startURL),
			zap.Duration("delay", delay),
			zap.Error(err),
		)
		return fmt.Errorf("failed to enqueue delayed crawl task: %w", err)
	}

	c.logger.Info("Enqueued delayed crawl task",
		zap.Uint("websiteID", websiteID),
		zap.String("url", startURL),
		zap.Duration("delay", delay),
		zap.String("taskID", info.ID),
	)

	return nil
}
