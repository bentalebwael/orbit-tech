package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/wbentaleb/student-report-service/internal/errors"
	"github.com/wbentaleb/student-report-service/internal/service"
)

type StudentReportHandler struct {
	reportService service.ReportService
	logger        *zap.Logger
}

func NewStudentReportHandler(reportService service.ReportService, logger *zap.Logger) *StudentReportHandler {
	return &StudentReportHandler{
		reportService: reportService,
		logger:        logger,
	}
}

func (h *StudentReportHandler) Handle(c *gin.Context) {
	studentID := c.Param("id")

	if err := validateStudentID(studentID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	pdfData, fileName, err := h.reportService.GenerateStudentReport(c.Request.Context(), studentID)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	c.Header("Content-Disposition", "attachment; filename="+fileName)
	c.Data(http.StatusOK, "application/pdf", pdfData)
}

func (h *StudentReportHandler) handleServiceError(c *gin.Context, err error) {
	switch {
	case errors.IsNotFound(err):
		c.JSON(http.StatusNotFound, gin.H{"error": "Student not found"})
	case errors.IsServiceError(err):
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Backend service unavailable"})
	case errors.IsPDFGenerationError(err):
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate PDF"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
	}
}
