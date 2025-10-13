package main

import (
	"fmt"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/wbentaleb/student-report-service/internal/client"
	"github.com/wbentaleb/student-report-service/internal/config"
	"github.com/wbentaleb/student-report-service/internal/handler"
	"github.com/wbentaleb/student-report-service/internal/service"
	"github.com/wbentaleb/student-report-service/pkg/logger"
	"go.uber.org/zap"
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
	backendClient := client.NewBackendClient(cfg.BackendURL, cfg.APIKey, log)
	pdfService := service.NewPDFService(log)

	// Initialize handlers
	healthHandler := handler.NewHealthHandler(backendClient)
	reportHandler := handler.NewReportHandler(backendClient, pdfService, log)

	// Setup Gin router
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()

	// Middleware for logging
	router.Use(func(c *gin.Context) {
		log.Info("Incoming request",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.String("remote_addr", c.ClientIP()))
		c.Next()
	})

	// Routes
	router.GET("/health", healthHandler.Handle)
	router.GET("/api/v1/students/:id/report", reportHandler.Handle)

	// Start server
	address := fmt.Sprintf(":%s", cfg.Port)
	log.Info("Server listening", zap.String("address", address))

	if err := router.Run(address); err != nil {
		log.Fatal("Failed to start server", zap.Error(err))
	}
}
