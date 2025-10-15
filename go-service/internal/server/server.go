package server

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/wbentaleb/student-report-service/internal/config"
	"github.com/wbentaleb/student-report-service/internal/handler"
	"github.com/wbentaleb/student-report-service/internal/middleware"
)

func NewRouter(
	cfg *config.Config,
	log *zap.Logger,
	healthHandler *handler.HealthHandler,
	reportHandler *handler.StudentReportHandler,
) *gin.Engine {

	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	applyMiddleware(router, cfg, log)
	defineRoutes(router, healthHandler, reportHandler)

	return router
}

func applyMiddleware(router *gin.Engine, cfg *config.Config, log *zap.Logger) {
	router.Use(gin.Recovery())
	router.Use(middleware.RequestID())
	router.Use(middleware.BasicSecurity())

	if cfg.EnableRateLimit {
		limiter := middleware.NewRateLimiter(cfg.RateLimitPerMinute)
		router.Use(limiter.Middleware())
		log.Info("Rate limiting enabled", zap.Int("requests_per_minute", cfg.RateLimitPerMinute))
	}

	router.Use(middleware.RequestLogger(log))
}

func defineRoutes(
	router *gin.Engine,
	healthHandler *handler.HealthHandler,
	reportHandler *handler.StudentReportHandler,
) {
	// Health check endpoint
	router.GET("/health", healthHandler.Handle)

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		v1.GET("/students/:id/report", reportHandler.Handle)
	}
}
