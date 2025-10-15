package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/wbentaleb/student-report-service/internal/cache"
	"github.com/wbentaleb/student-report-service/internal/config"
	"github.com/wbentaleb/student-report-service/internal/external"
	"github.com/wbentaleb/student-report-service/internal/handler"
	"github.com/wbentaleb/student-report-service/internal/middleware"
	"github.com/wbentaleb/student-report-service/internal/service"
	"github.com/wbentaleb/student-report-service/pkg/logger"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	log, err := logger.New(cfg.Environment, cfg.LogLevel)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer log.Sync()

	log.Info("Starting student report service",
		zap.String("environment", cfg.Environment),
		zap.String("port", cfg.Port),
		zap.String("backend_url", cfg.BackendURL))

	// Initialize clients and services
	backendClient := external.NewBackendClient(cfg.BackendURL, cfg.APIKey, log)
	pdfService := service.NewPDFService(log)

	// Initialize file cache
	var pdfCache *cache.FileCache
	if cfg.EnableCache {
		var err error
		pdfCache, err = cache.NewFileCache(cfg.CachePath, cfg.CacheTTL)
		if err != nil {
			log.Warn("Failed to initialize cache, continuing without cache", zap.Error(err))
		} else {
			log.Info("Cache initialized",
				zap.String("path", cfg.CachePath),
				zap.Duration("ttl", cfg.CacheTTL))
		}
	}

	// Initialize handlers
	healthHandler := handler.NewHealthHandler(backendClient)
	reportHandler := handler.NewReportHandler(backendClient, pdfService, pdfCache, log)

	// Setup Gin router
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	// Basic middleware stack
	router.Use(gin.Recovery())
	router.Use(middleware.RequestID())
	router.Use(middleware.BasicSecurity())

	// Optional rate limiting
	if cfg.EnableRateLimit {
		limiter := middleware.NewRateLimiter(cfg.RateLimitPerMinute)
		router.Use(limiter.Middleware())
		log.Info("Rate limiting enabled", zap.Int("requests_per_minute", cfg.RateLimitPerMinute))
	}

	// Request logging middleware
	router.Use(func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path

		c.Next()

		duration := time.Since(start)
		log.Info("Request completed",
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.Int("status", c.Writer.Status()),
			zap.Duration("duration", duration),
			zap.String("remote_addr", c.ClientIP()))
	})

	// Routes
	router.GET("/health", healthHandler.Handle)
	router.GET("/api/v1/students/:id/report", reportHandler.Handle)

	// Server with graceful shutdown
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Port),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	// Start server in goroutine
	go func() {
		log.Info("Starting server", zap.String("address", srv.Addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error("Server forced to shutdown", zap.Error(err))
	}

	log.Info("Server exited")
}
