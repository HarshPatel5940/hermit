package middlewares

import (
	"net/http"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// RateLimiterConfig holds rate limiter configuration.
type RateLimiterConfig struct {
	RequestsPerMinute int64
	Burst             int64
	Enabled           bool
}

// ipRateLimiter tracks requests per IP.
type ipRateLimiter struct {
	visitors map[string]*visitor
	mu       sync.RWMutex
	limit    int64
	window   time.Duration
}

// visitor tracks requests for a single IP.
type visitor struct {
	requests  int64
	firstSeen time.Time
}

// NewRateLimiter creates a new rate limiter middleware.
func NewRateLimiter(cfg RateLimiterConfig, logger *zap.Logger) echo.MiddlewareFunc {
	if !cfg.Enabled {
		logger.Info("Rate limiting disabled")
		return func(next echo.HandlerFunc) echo.HandlerFunc {
			return next
		}
	}

	limiter := &ipRateLimiter{
		visitors: make(map[string]*visitor),
		limit:    cfg.RequestsPerMinute,
		window:   time.Minute,
	}

	// Cleanup old visitors every 5 minutes
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			limiter.cleanup()
		}
	}()

	logger.Info("Rate limiting enabled",
		zap.Int64("requests_per_minute", cfg.RequestsPerMinute),
	)

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ip := c.RealIP()

			if !limiter.allow(ip) {
				logger.Warn("Rate limit exceeded",
					zap.String("ip", ip),
					zap.String("path", c.Request().URL.Path),
				)
				return c.JSON(http.StatusTooManyRequests, map[string]string{
					"error":   "Rate limit exceeded",
					"message": "Too many requests. Please try again later.",
				})
			}

			return next(c)
		}
	}
}

// allow checks if the IP is allowed to make a request.
func (l *ipRateLimiter) allow(ip string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	v, exists := l.visitors[ip]

	if !exists || now.Sub(v.firstSeen) > l.window {
		l.visitors[ip] = &visitor{
			requests:  1,
			firstSeen: now,
		}
		return true
	}

	if v.requests >= l.limit {
		return false
	}

	v.requests++
	return true
}

// cleanup removes old visitor entries.
func (l *ipRateLimiter) cleanup() {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	for ip, v := range l.visitors {
		if now.Sub(v.firstSeen) > l.window*2 {
			delete(l.visitors, ip)
		}
	}
}
