package middlewares

import (
	"hermit/internal/config"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"
)

func SetupMiddlewares(e *echo.Echo, logger *zap.Logger, cfg *config.Config) {
	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogURI:    true,
		LogStatus: true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			logger.Info("request",
				zap.String("URI", v.URI),
				zap.Int("status", v.Status),
				zap.String("method", v.Method),
				zap.Duration("latency", v.Latency),
			)
			return nil
		},
	}))

	e.Use(middleware.Recover())
	e.Use(middleware.RemoveTrailingSlash())
	e.Use(middleware.Decompress())

	// Apply security headers
	e.Use(middleware.SecureWithConfig(middleware.SecureConfig{
		XSSProtection:         "1; mode=block",
		ContentTypeNosniff:    "nosniff",
		XFrameOptions:         "DENY",
		HSTSMaxAge:            31536000,
		ContentSecurityPolicy: "default-src 'self'",
	}))

	// Apply custom rate limiter
	rateLimiterCfg := RateLimiterConfig{
		RequestsPerMinute: cfg.RateLimitRequestsPerMin,
		Burst:             cfg.RateLimitBurst,
		Enabled:           cfg.RateLimitEnabled,
	}
	e.Use(NewRateLimiter(rateLimiterCfg, logger))

	// CORS configuration
	corsOrigins := []string{"*"}
	if cfg.Port == "8080" { // Production check - adjust as needed
		corsOrigins = []string{"https://yourdomain.com"} // Configure via env in production
	}

	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     corsOrigins,
		AllowMethods:     []string{http.MethodGet, http.MethodPost, http.MethodPatch, http.MethodPut, http.MethodDelete, http.MethodOptions},
		AllowHeaders:     []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization},
		AllowCredentials: true,
		MaxAge:           3600,
	}))
}
