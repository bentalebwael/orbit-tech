package handler

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/wbentaleb/student-report-service/internal/cache"
	"github.com/wbentaleb/student-report-service/internal/dto"
	"github.com/wbentaleb/student-report-service/internal/errors"
	"github.com/wbentaleb/student-report-service/internal/external"
	"github.com/wbentaleb/student-report-service/internal/service"
)

// getRequestID retrieves the request ID from the context
func getRequestID(c *gin.Context) string {
	if requestID, exists := c.Get("RequestID"); exists {
		if id, ok := requestID.(string); ok {
			return id
		}
	}
	return ""
}

// ReportHandler handles PDF report generation requests
type ReportHandler struct {
	backendClient external.BackendService
	pdfService    service.PDFGenerator
	cache         *cache.FileCache
	logger        *zap.Logger
}

// NewReportHandler creates a new report handler
func NewReportHandler(backendClient external.BackendService, pdfService service.PDFGenerator, cache *cache.FileCache, logger *zap.Logger) *ReportHandler {
	return &ReportHandler{
		backendClient: backendClient,
		pdfService:    pdfService,
		cache:         cache,
		logger:        logger,
	}
}

// Handle processes PDF report generation requests
func (h *ReportHandler) Handle(c *gin.Context) {
	studentID := c.Param("id")

	if err := validateStudentID(studentID); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:     err.Error(),
			RequestID: getRequestID(c),
		})
		return
	}

	h.logger.Info("Received report generation request",
		zap.String("student_id", studentID),
		zap.String("remote_addr", c.ClientIP()))

	// Fetch student data from backend
	student, err := h.backendClient.GetStudent(c.Request.Context(), studentID)
	if err != nil {
		switch {
		case errors.IsNotFound(err):
			h.logger.Warn("Student not found",
				zap.String("student_id", studentID))
			c.JSON(http.StatusNotFound, dto.ErrorResponse{
				Error:     "Student not found",
				RequestID: getRequestID(c),
			})
		case errors.IsServiceError(err):
			h.logger.Error("Backend service error",
				zap.String("student_id", studentID),
				zap.Error(err))
			c.JSON(http.StatusServiceUnavailable, dto.ErrorResponse{
				Error:     "Backend service unavailable",
				RequestID: getRequestID(c),
			})
		default:
			h.logger.Error("Unexpected error fetching student",
				zap.String("student_id", studentID),
				zap.Error(err))
			c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
				Error:     "Internal server error",
				RequestID: getRequestID(c),
			})
		}
		return
	}

	// Generate hash from current student data
	hash := cache.GenerateStudentHash(student)

	// Check cache using student ID and hash
	if h.cache != nil {
		if pdfData, found := h.cache.Get(studentID, hash); found {
			h.logger.Info("Cache hit",
				zap.String("student_id", studentID),
				zap.String("hash", hash))

			// Set response headers
			timestamp := time.Now().Unix()
			filename := fmt.Sprintf("student_%s_report_%d.pdf", studentID, timestamp)

			c.Header("Content-Type", "application/pdf")
			c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
			c.Header("X-Cache", "HIT")
			c.Header("Content-Length", fmt.Sprintf("%d", len(pdfData)))

			c.Data(http.StatusOK, "application/pdf", pdfData)
			return
		}
	}

	// Cache miss - generate new PDF
	h.logger.Info("Cache miss, generating PDF",
		zap.String("student_id", studentID),
		zap.String("hash", hash))

	pdfBytes, err := h.pdfService.GenerateStudentReport(student)
	if err != nil {
		h.logger.Error("Failed to generate PDF",
			zap.String("student_id", studentID),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:     "Failed to generate PDF report",
			RequestID: getRequestID(c),
		})
		return
	}

	// Store in cache with hash
	if h.cache != nil {
		if err := h.cache.Set(studentID, pdfBytes, hash); err != nil {
			h.logger.Warn("Failed to cache PDF", zap.Error(err))
			// Continue anyway - caching is not critical
		}
	}

	// Set response headers
	timestamp := time.Now().Unix()
	filename := fmt.Sprintf("student_%s_report_%d.pdf", studentID, timestamp)

	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Header("X-Cache", "MISS")
	c.Header("Content-Length", fmt.Sprintf("%d", len(pdfBytes)))

	h.logger.Info("PDF generated successfully",
		zap.String("student_id", studentID),
		zap.Int("pdf_size", len(pdfBytes)))

	c.Data(http.StatusOK, "application/pdf", pdfBytes)
}
