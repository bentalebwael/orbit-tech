package handler

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	serviceErrors "github.com/wbentaleb/student-report-service/internal/errors"
)

// Mock ReportService
type MockReportService struct {
	mock.Mock
}

func (m *MockReportService) GenerateStudentReport(ctx context.Context, studentID string) (pdfData []byte, fileName string, err error) {
	args := m.Called(ctx, studentID)
	if args.Get(0) == nil {
		return nil, args.String(1), args.Error(2)
	}
	return args.Get(0).([]byte), args.String(1), args.Error(2)
}

func setupTestRouter(handler *StudentReportHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/api/v1/students/:id/report", handler.Handle)
	return router
}

func TestNewStudentReportHandler(t *testing.T) {
	logger := zap.NewNop()
	mockService := new(MockReportService)

	handler := NewStudentReportHandler(mockService, logger)

	assert.NotNil(t, handler)
	assert.Equal(t, mockService, handler.reportService)
	assert.Equal(t, logger, handler.logger)
}

func TestHandle_Success(t *testing.T) {
	// Setup
	logger := zap.NewNop()
	mockService := new(MockReportService)
	handler := NewStudentReportHandler(mockService, logger)
	router := setupTestRouter(handler)

	studentID := "12345"
	pdfData := []byte("test pdf content")
	fileName := "student_12345_report.pdf"

	// Setup mock
	mockService.On("GenerateStudentReport", mock.Anything, studentID).Return(pdfData, fileName, nil)

	// Create request
	req, _ := http.NewRequest("GET", "/api/v1/students/12345/report", nil)
	rec := httptest.NewRecorder()

	// Execute
	router.ServeHTTP(rec, req)

	// Assert
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "application/pdf", rec.Header().Get("Content-Type"))
	assert.Equal(t, "attachment; filename=student_12345_report.pdf", rec.Header().Get("Content-Disposition"))
	assert.Equal(t, pdfData, rec.Body.Bytes())

	mockService.AssertExpectations(t)
}

func TestHandle_InvalidStudentID_Empty(t *testing.T) {
	// Setup
	logger := zap.NewNop()
	mockService := new(MockReportService)
	handler := NewStudentReportHandler(mockService, logger)

	// Create gin context with empty ID
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	ginCtx, _ := gin.CreateTestContext(rec)
	req, _ := http.NewRequest("GET", "/api/v1/students//report", nil)
	ginCtx.Request = req
	ginCtx.Params = []gin.Param{{Key: "id", Value: ""}}

	// Execute
	handler.Handle(ginCtx)

	// Assert
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "student ID cannot be empty")

	mockService.AssertNotCalled(t, "GenerateStudentReport")
}

func TestHandle_InvalidStudentID_NonNumeric(t *testing.T) {
	// Setup
	logger := zap.NewNop()
	mockService := new(MockReportService)
	handler := NewStudentReportHandler(mockService, logger)
	router := setupTestRouter(handler)

	// Create request with non-numeric ID
	req, _ := http.NewRequest("GET", "/api/v1/students/abc123/report", nil)
	rec := httptest.NewRecorder()

	// Execute
	router.ServeHTTP(rec, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "student ID must be numeric (1-20 digits)")

	mockService.AssertNotCalled(t, "GenerateStudentReport")
}

func TestHandle_InvalidStudentID_TooLong(t *testing.T) {
	// Setup
	logger := zap.NewNop()
	mockService := new(MockReportService)
	handler := NewStudentReportHandler(mockService, logger)
	router := setupTestRouter(handler)

	// Create request with ID longer than 20 digits
	req, _ := http.NewRequest("GET", "/api/v1/students/123456789012345678901/report", nil)
	rec := httptest.NewRecorder()

	// Execute
	router.ServeHTTP(rec, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "student ID must be numeric (1-20 digits)")

	mockService.AssertNotCalled(t, "GenerateStudentReport")
}

func TestHandle_StudentNotFound(t *testing.T) {
	// Setup
	logger := zap.NewNop()
	mockService := new(MockReportService)
	handler := NewStudentReportHandler(mockService, logger)
	router := setupTestRouter(handler)

	studentID := "99999"
	notFoundErr := &serviceErrors.NotFoundError{Resource: "Student"}

	// Setup mock
	mockService.On("GenerateStudentReport", mock.Anything, studentID).Return(nil, "", notFoundErr)

	// Create request
	req, _ := http.NewRequest("GET", "/api/v1/students/99999/report", nil)
	rec := httptest.NewRecorder()

	// Execute
	router.ServeHTTP(rec, req)

	// Assert
	assert.Equal(t, http.StatusNotFound, rec.Code)
	assert.Contains(t, rec.Body.String(), "Student not found")

	mockService.AssertExpectations(t)
}

func TestHandle_ServiceError(t *testing.T) {
	// Setup
	logger := zap.NewNop()
	mockService := new(MockReportService)
	handler := NewStudentReportHandler(mockService, logger)
	router := setupTestRouter(handler)

	studentID := "12345"
	serviceErr := &serviceErrors.ServiceError{Service: "Backend", Err: errors.New("backend unavailable")}

	// Setup mock
	mockService.On("GenerateStudentReport", mock.Anything, studentID).Return(nil, "", serviceErr)

	// Create request
	req, _ := http.NewRequest("GET", "/api/v1/students/12345/report", nil)
	rec := httptest.NewRecorder()

	// Execute
	router.ServeHTTP(rec, req)

	// Assert
	assert.Equal(t, http.StatusServiceUnavailable, rec.Code)
	assert.Contains(t, rec.Body.String(), "Backend service unavailable")

	mockService.AssertExpectations(t)
}

func TestHandle_PDFGenerationError(t *testing.T) {
	// Setup
	logger := zap.NewNop()
	mockService := new(MockReportService)
	handler := NewStudentReportHandler(mockService, logger)
	router := setupTestRouter(handler)

	studentID := "12345"
	pdfErr := serviceErrors.NewPDFGenerationError(errors.New("pdf generation failed"))

	// Setup mock
	mockService.On("GenerateStudentReport", mock.Anything, studentID).Return(nil, "", pdfErr)

	// Create request
	req, _ := http.NewRequest("GET", "/api/v1/students/12345/report", nil)
	rec := httptest.NewRecorder()

	// Execute
	router.ServeHTTP(rec, req)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.Contains(t, rec.Body.String(), "Failed to generate PDF")

	mockService.AssertExpectations(t)
}

func TestHandle_GenericError(t *testing.T) {
	// Setup
	logger := zap.NewNop()
	mockService := new(MockReportService)
	handler := NewStudentReportHandler(mockService, logger)
	router := setupTestRouter(handler)

	studentID := "12345"
	genericErr := errors.New("unexpected error")

	// Setup mock
	mockService.On("GenerateStudentReport", mock.Anything, studentID).Return(nil, "", genericErr)

	// Create request
	req, _ := http.NewRequest("GET", "/api/v1/students/12345/report", nil)
	rec := httptest.NewRecorder()

	// Execute
	router.ServeHTTP(rec, req)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.Contains(t, rec.Body.String(), "Internal server error")

	mockService.AssertExpectations(t)
}

func TestHandle_ValidStudentIDs(t *testing.T) {
	// Setup
	logger := zap.NewNop()
	mockService := new(MockReportService)
	handler := NewStudentReportHandler(mockService, logger)
	router := setupTestRouter(handler)

	testCases := []struct {
		name      string
		studentID string
		valid     bool
	}{
		{
			name:      "single digit",
			studentID: "1",
			valid:     true,
		},
		{
			name:      "multiple digits",
			studentID: "12345",
			valid:     true,
		},
		{
			name:      "maximum length (20 digits)",
			studentID: "12345678901234567890",
			valid:     true,
		},
		{
			name:      "with leading zeros",
			studentID: "00123",
			valid:     true,
		},
		{
			name:      "contains letter",
			studentID: "123a45",
			valid:     false,
		},
		{
			name:      "contains hyphen",
			studentID: "123-45",
			valid:     false,
		},
		{
			name:      "contains space",
			studentID: "123 45",
			valid:     false,
		},
		{
			name:      "too long (21 digits)",
			studentID: "123456789012345678901",
			valid:     false,
		},
		{
			name:      "empty",
			studentID: "",
			valid:     false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.valid {
				// Setup mock for valid IDs
				pdfData := []byte("test pdf")
				fileName := "test.pdf"
				mockService.On("GenerateStudentReport", mock.Anything, tc.studentID).Return(pdfData, fileName, nil).Maybe()
			}

			// Create request
			req, _ := http.NewRequest("GET", "/api/v1/students/"+tc.studentID+"/report", nil)
			rec := httptest.NewRecorder()

			// Execute
			router.ServeHTTP(rec, req)

			// Assert
			if tc.valid {
				assert.Equal(t, http.StatusOK, rec.Code)
			} else {
				assert.NotEqual(t, http.StatusOK, rec.Code)
			}
		})
	}
}

func TestHandle_LargePDFResponse(t *testing.T) {
	// Setup
	logger := zap.NewNop()
	mockService := new(MockReportService)
	handler := NewStudentReportHandler(mockService, logger)
	router := setupTestRouter(handler)

	studentID := "12345"
	// Create a large PDF (1MB)
	largePDF := make([]byte, 1024*1024)
	for i := range largePDF {
		largePDF[i] = byte(i % 256)
	}
	fileName := "large_report.pdf"

	// Setup mock
	mockService.On("GenerateStudentReport", mock.Anything, studentID).Return(largePDF, fileName, nil)

	// Create request
	req, _ := http.NewRequest("GET", "/api/v1/students/12345/report", nil)
	rec := httptest.NewRecorder()

	// Execute
	router.ServeHTTP(rec, req)

	// Assert
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, largePDF, rec.Body.Bytes())
	assert.Equal(t, "application/pdf", rec.Header().Get("Content-Type"))

	mockService.AssertExpectations(t)
}

func TestHandle_ContextCancellation(t *testing.T) {
	// Setup
	logger := zap.NewNop()
	mockService := new(MockReportService)
	handler := NewStudentReportHandler(mockService, logger)

	studentID := "12345"

	// Setup mock that checks context
	mockService.On("GenerateStudentReport", mock.Anything, studentID).Run(func(args mock.Arguments) {
		ctx := args.Get(0).(context.Context)
		require.NotNil(t, ctx)
	}).Return([]byte("pdf"), "file.pdf", nil)

	// Create request with context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	req, _ := http.NewRequestWithContext(ctx, "GET", "/api/v1/students/12345/report", nil)
	rec := httptest.NewRecorder()

	// Create gin context
	gin.SetMode(gin.TestMode)
	ginCtx, _ := gin.CreateTestContext(rec)
	ginCtx.Request = req
	ginCtx.Params = []gin.Param{{Key: "id", Value: studentID}}

	// Execute
	handler.Handle(ginCtx)

	// Assert
	assert.Equal(t, http.StatusOK, rec.Code)
	mockService.AssertExpectations(t)
}

func TestHandle_EmptyPDFData(t *testing.T) {
	// Setup
	logger := zap.NewNop()
	mockService := new(MockReportService)
	handler := NewStudentReportHandler(mockService, logger)
	router := setupTestRouter(handler)

	studentID := "12345"
	emptyPDF := []byte{}
	fileName := "empty.pdf"

	// Setup mock
	mockService.On("GenerateStudentReport", mock.Anything, studentID).Return(emptyPDF, fileName, nil)

	// Create request
	req, _ := http.NewRequest("GET", "/api/v1/students/12345/report", nil)
	rec := httptest.NewRecorder()

	// Execute
	router.ServeHTTP(rec, req)

	// Assert
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Len(t, rec.Body.Bytes(), 0)
	assert.Equal(t, "application/pdf", rec.Header().Get("Content-Type"))
	assert.Equal(t, "attachment; filename=empty.pdf", rec.Header().Get("Content-Disposition"))

	mockService.AssertExpectations(t)
}

func TestHandle_SpecialCharactersInFileName(t *testing.T) {
	// Setup
	logger := zap.NewNop()
	mockService := new(MockReportService)
	handler := NewStudentReportHandler(mockService, logger)
	router := setupTestRouter(handler)

	studentID := "12345"
	pdfData := []byte("test pdf")
	// File name with special characters
	fileName := "student_12345_report's & \"quotes\".pdf"

	// Setup mock
	mockService.On("GenerateStudentReport", mock.Anything, studentID).Return(pdfData, fileName, nil)

	// Create request
	req, _ := http.NewRequest("GET", "/api/v1/students/12345/report", nil)
	rec := httptest.NewRecorder()

	// Execute
	router.ServeHTTP(rec, req)

	// Assert
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, pdfData, rec.Body.Bytes())
	// Check that filename is properly escaped in Content-Disposition
	assert.Contains(t, rec.Header().Get("Content-Disposition"), "filename=")

	mockService.AssertExpectations(t)
}

// Benchmark tests
func BenchmarkHandle_Success(b *testing.B) {
	logger := zap.NewNop()
	mockService := new(MockReportService)
	handler := NewStudentReportHandler(mockService, logger)

	pdfData := bytes.Repeat([]byte("test"), 250) // 1KB PDF
	fileName := "report.pdf"

	mockService.On("GenerateStudentReport", mock.Anything, "12345").Return(pdfData, fileName, nil)

	gin.SetMode(gin.ReleaseMode)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("GET", "/api/v1/students/12345/report", nil)
		rec := httptest.NewRecorder()

		ginCtx, _ := gin.CreateTestContext(rec)
		ginCtx.Request = req
		ginCtx.Params = []gin.Param{{Key: "id", Value: "12345"}}

		handler.Handle(ginCtx)
	}
}

func BenchmarkStudentIDValidation(b *testing.B) {
	testID := "12345678901234567890" // 20 digits

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = validateStudentID(testID)
	}
}
