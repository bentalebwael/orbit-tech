package handler

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/wbentaleb/student-report-service/internal/client"
	"github.com/wbentaleb/student-report-service/internal/service"
	"go.uber.org/zap"
)

// ReportHandler handles PDF report generation requests
type ReportHandler struct {
	backendClient *client.BackendClient
	pdfService    *service.PDFService
	logger        *zap.Logger
}

// NewReportHandler creates a new report handler
func NewReportHandler(backendClient *client.BackendClient, pdfService *service.PDFService, logger *zap.Logger) *ReportHandler {
	return &ReportHandler{
		backendClient: backendClient,
		pdfService:    pdfService,
		logger:        logger,
	}
}

// Handle processes PDF report generation requests
func (h *ReportHandler) Handle(c *gin.Context) {
	studentID := c.Param("id")

	h.logger.Info("Received report generation request",
		zap.String("student_id", studentID),
		zap.String("remote_addr", c.ClientIP()))

	// Fetch student data from backend
	student, err := h.backendClient.GetStudent(c.Request.Context(), studentID)
	if err != nil {
		h.logger.Error("Failed to fetch student data",
			zap.String("student_id", studentID),
			zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{
			"error":  "Student not found or backend service unavailable",
			"status": http.StatusNotFound,
		})
		return
	}

	// Generate PDF
	pdfBytes, err := h.pdfService.GenerateStudentReport(student)
	if err != nil {
		h.logger.Error("Failed to generate PDF",
			zap.String("student_id", studentID),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":  "Failed to generate PDF report",
			"status": http.StatusInternalServerError,
		})
		return
	}

	// Set response headers
	timestamp := time.Now().Unix()
	filename := fmt.Sprintf("student_%s_report_%d.pdf", studentID, timestamp)

	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Header("Cache-Control", "no-cache")
	c.Header("Content-Length", fmt.Sprintf("%d", len(pdfBytes)))

	h.logger.Info("PDF generated successfully",
		zap.String("student_id", studentID),
		zap.Int("pdf_size", len(pdfBytes)))

	c.Data(http.StatusOK, "application/pdf", pdfBytes)
}
