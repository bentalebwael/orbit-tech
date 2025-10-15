package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/wbentaleb/student-report-service/internal/cache"
	"github.com/wbentaleb/student-report-service/internal/dto"
	serviceErrors "github.com/wbentaleb/student-report-service/internal/errors"
)

// Mock implementations
type MockBackendService struct {
	mock.Mock
}

func (m *MockBackendService) GetStudent(ctx context.Context, id string) (*dto.Student, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.Student), args.Error(1)
}

func (m *MockBackendService) CheckHealth(ctx context.Context) bool {
	args := m.Called(ctx)
	return args.Bool(0)
}

type MockPDFGenerator struct {
	mock.Mock
}

func (m *MockPDFGenerator) GenerateStudentReport(student *dto.Student) ([]byte, error) {
	args := m.Called(student)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]byte), args.Error(1)
}

type MockPDFCache struct {
	mock.Mock
}

func (m *MockPDFCache) Get(studentID, hash string) ([]byte, bool) {
	args := m.Called(studentID, hash)
	if args.Get(0) == nil {
		return nil, args.Bool(1)
	}
	return args.Get(0).([]byte), args.Bool(1)
}

func (m *MockPDFCache) Set(studentID string, data []byte, hash string) error {
	args := m.Called(studentID, data, hash)
	return args.Error(0)
}

// Test helper functions
func createTestStudent() *dto.Student {
	return &dto.Student{
		ID:                 12345,
		Name:               "John Doe",
		Email:              "john.doe@example.com",
		SystemAccess:       true,
		Phone:              "1234567890",
		Gender:             "Male",
		DOB:                "2000-01-01T00:00:00Z",
		Class:              "10",
		Section:            "A",
		Roll:               15,
		CurrentAddress:     "123 Main St",
		PermanentAddress:   "456 Oak Ave",
		FatherName:         "Robert Doe",
		FatherPhone:        "9876543210",
		MotherName:         "Jane Doe",
		MotherPhone:        "9876543211",
		GuardianName:       "Uncle Bob",
		GuardianPhone:      "9876543212",
		RelationOfGuardian: "Uncle",
		AdmissionDate:      "2015-06-01T00:00:00Z",
		ReporterName:       "Admin",
		LastUpdated:        "2024-01-01T10:00:00Z",
	}
}

func TestNewStudentReportService(t *testing.T) {
	logger := zap.NewNop()
	mockBackend := new(MockBackendService)
	mockPDFGen := new(MockPDFGenerator)
	mockCache := new(MockPDFCache)

	service := NewStudentReportService(mockBackend, mockPDFGen, mockCache, logger)

	assert.NotNil(t, service)
	assert.Equal(t, mockBackend, service.backendClient)
	assert.Equal(t, mockPDFGen, service.pdfGenerator)
	assert.Equal(t, mockCache, service.pdfCache)
	assert.Equal(t, logger, service.logger)
}

func TestGenerateStudentReport_Success_WithCache(t *testing.T) {
	// Setup
	logger := zap.NewNop()
	mockBackend := new(MockBackendService)
	mockPDFGen := new(MockPDFGenerator)
	mockCache := new(MockPDFCache)

	service := NewStudentReportService(mockBackend, mockPDFGen, mockCache, logger)

	ctx := context.Background()
	studentID := "12345"
	student := createTestStudent()
	cachedPDF := []byte("cached pdf content")

	// Setup mocks
	mockBackend.On("GetStudent", ctx, studentID).Return(student, nil)

	// Calculate the expected hash
	contentHash := cache.GenerateStudentHash(student)
	mockCache.On("Get", studentID, contentHash).Return(cachedPDF, true)

	// Execute
	pdfData, fileName, err := service.GenerateStudentReport(ctx, studentID)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, cachedPDF, pdfData)
	assert.Equal(t, "student_12345_report.pdf", fileName)

	// Verify mock expectations
	mockBackend.AssertExpectations(t)
	mockCache.AssertExpectations(t)
	mockPDFGen.AssertNotCalled(t, "GenerateStudentReport")
}

func TestGenerateStudentReport_Success_WithoutCache(t *testing.T) {
	// Setup
	logger := zap.NewNop()
	mockBackend := new(MockBackendService)
	mockPDFGen := new(MockPDFGenerator)
	mockCache := new(MockPDFCache)

	service := NewStudentReportService(mockBackend, mockPDFGen, mockCache, logger)

	ctx := context.Background()
	studentID := "12345"
	student := createTestStudent()
	generatedPDF := []byte("generated pdf content")

	// Setup mocks
	mockBackend.On("GetStudent", ctx, studentID).Return(student, nil)

	// Calculate the expected hash
	contentHash := cache.GenerateStudentHash(student)
	mockCache.On("Get", studentID, contentHash).Return(nil, false)
	mockPDFGen.On("GenerateStudentReport", student).Return(generatedPDF, nil)
	mockCache.On("Set", studentID, generatedPDF, contentHash).Return(nil)

	// Execute
	pdfData, fileName, err := service.GenerateStudentReport(ctx, studentID)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, generatedPDF, pdfData)
	assert.Equal(t, "student_12345_report.pdf", fileName)

	// Verify mock expectations
	mockBackend.AssertExpectations(t)
	mockCache.AssertExpectations(t)
	mockPDFGen.AssertExpectations(t)
}

func TestGenerateStudentReport_Success_NilCache(t *testing.T) {
	// Setup
	logger := zap.NewNop()
	mockBackend := new(MockBackendService)
	mockPDFGen := new(MockPDFGenerator)

	// Create service without cache
	service := NewStudentReportService(mockBackend, mockPDFGen, nil, logger)

	ctx := context.Background()
	studentID := "12345"
	student := createTestStudent()
	generatedPDF := []byte("generated pdf content")

	// Setup mocks
	mockBackend.On("GetStudent", ctx, studentID).Return(student, nil)
	mockPDFGen.On("GenerateStudentReport", student).Return(generatedPDF, nil)

	// Execute
	pdfData, fileName, err := service.GenerateStudentReport(ctx, studentID)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, generatedPDF, pdfData)
	assert.Equal(t, "student_12345_report.pdf", fileName)

	// Verify mock expectations
	mockBackend.AssertExpectations(t)
	mockPDFGen.AssertExpectations(t)
}

func TestGenerateStudentReport_BackendError(t *testing.T) {
	// Setup
	logger := zap.NewNop()
	mockBackend := new(MockBackendService)
	mockPDFGen := new(MockPDFGenerator)
	mockCache := new(MockPDFCache)

	service := NewStudentReportService(mockBackend, mockPDFGen, mockCache, logger)

	ctx := context.Background()
	studentID := "12345"
	backendErr := errors.New("backend service error")

	// Setup mocks
	mockBackend.On("GetStudent", ctx, studentID).Return(nil, backendErr)

	// Execute
	pdfData, fileName, err := service.GenerateStudentReport(ctx, studentID)

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to fetch student data")
	assert.Nil(t, pdfData)
	assert.Empty(t, fileName)

	// Verify mock expectations
	mockBackend.AssertExpectations(t)
	mockCache.AssertNotCalled(t, "Get")
	mockPDFGen.AssertNotCalled(t, "GenerateStudentReport")
}

func TestGenerateStudentReport_PDFGenerationError(t *testing.T) {
	// Setup
	logger := zap.NewNop()
	mockBackend := new(MockBackendService)
	mockPDFGen := new(MockPDFGenerator)
	mockCache := new(MockPDFCache)

	service := NewStudentReportService(mockBackend, mockPDFGen, mockCache, logger)

	ctx := context.Background()
	studentID := "12345"
	student := createTestStudent()
	pdfGenErr := errors.New("pdf generation failed")

	// Setup mocks
	mockBackend.On("GetStudent", ctx, studentID).Return(student, nil)

	// Calculate the expected hash
	contentHash := cache.GenerateStudentHash(student)
	mockCache.On("Get", studentID, contentHash).Return(nil, false)
	mockPDFGen.On("GenerateStudentReport", student).Return(nil, pdfGenErr)

	// Execute
	pdfData, fileName, err := service.GenerateStudentReport(ctx, studentID)

	// Assert
	require.Error(t, err)
	assert.IsType(t, &serviceErrors.PDFGenerationError{}, err)
	assert.Nil(t, pdfData)
	assert.Empty(t, fileName)

	// Verify mock expectations
	mockBackend.AssertExpectations(t)
	mockCache.AssertExpectations(t)
	mockPDFGen.AssertExpectations(t)
}

func TestGenerateStudentReport_CacheSetError(t *testing.T) {
	// Setup
	logger := zap.NewNop()
	mockBackend := new(MockBackendService)
	mockPDFGen := new(MockPDFGenerator)
	mockCache := new(MockPDFCache)

	service := NewStudentReportService(mockBackend, mockPDFGen, mockCache, logger)

	ctx := context.Background()
	studentID := "12345"
	student := createTestStudent()
	generatedPDF := []byte("generated pdf content")
	cacheErr := errors.New("cache write failed")

	// Setup mocks
	mockBackend.On("GetStudent", ctx, studentID).Return(student, nil)

	// Calculate the expected hash
	contentHash := cache.GenerateStudentHash(student)
	mockCache.On("Get", studentID, contentHash).Return(nil, false)
	mockPDFGen.On("GenerateStudentReport", student).Return(generatedPDF, nil)
	mockCache.On("Set", studentID, generatedPDF, contentHash).Return(cacheErr)

	// Execute
	pdfData, fileName, err := service.GenerateStudentReport(ctx, studentID)

	// Assert - should still succeed despite cache error
	require.NoError(t, err)
	assert.Equal(t, generatedPDF, pdfData)
	assert.Equal(t, "student_12345_report.pdf", fileName)

	// Verify mock expectations
	mockBackend.AssertExpectations(t)
	mockCache.AssertExpectations(t)
	mockPDFGen.AssertExpectations(t)
}

func TestFetchStudentData_Success(t *testing.T) {
	// Setup
	logger := zap.NewNop()
	mockBackend := new(MockBackendService)
	mockPDFGen := new(MockPDFGenerator)
	mockCache := new(MockPDFCache)

	service := NewStudentReportService(mockBackend, mockPDFGen, mockCache, logger)

	ctx := context.Background()
	studentID := "12345"
	expectedStudent := createTestStudent()

	// Setup mocks
	mockBackend.On("GetStudent", ctx, studentID).Return(expectedStudent, nil)

	// Execute
	student, err := service.fetchStudentData(ctx, studentID)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, expectedStudent, student)

	// Verify mock expectations
	mockBackend.AssertExpectations(t)
}

func TestFetchStudentData_Error(t *testing.T) {
	// Setup
	logger := zap.NewNop()
	mockBackend := new(MockBackendService)
	mockPDFGen := new(MockPDFGenerator)
	mockCache := new(MockPDFCache)

	service := NewStudentReportService(mockBackend, mockPDFGen, mockCache, logger)

	ctx := context.Background()
	studentID := "12345"
	backendErr := errors.New("backend error")

	// Setup mocks
	mockBackend.On("GetStudent", ctx, studentID).Return(nil, backendErr)

	// Execute
	student, err := service.fetchStudentData(ctx, studentID)

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to fetch student data")
	assert.Nil(t, student)

	// Verify mock expectations
	mockBackend.AssertExpectations(t)
}

func TestTryGetFromCache_Found(t *testing.T) {
	// Setup
	logger := zap.NewNop()
	mockBackend := new(MockBackendService)
	mockPDFGen := new(MockPDFGenerator)
	mockCache := new(MockPDFCache)

	service := NewStudentReportService(mockBackend, mockPDFGen, mockCache, logger)

	studentID := "12345"
	contentHash := "abcd1234"
	cachedPDF := []byte("cached pdf content")

	// Setup mocks
	mockCache.On("Get", studentID, contentHash).Return(cachedPDF, true)

	// Execute
	result := service.tryGetFromCache(studentID, contentHash)

	// Assert
	assert.Equal(t, cachedPDF, result)

	// Verify mock expectations
	mockCache.AssertExpectations(t)
}

func TestTryGetFromCache_NotFound(t *testing.T) {
	// Setup
	logger := zap.NewNop()
	mockBackend := new(MockBackendService)
	mockPDFGen := new(MockPDFGenerator)
	mockCache := new(MockPDFCache)

	service := NewStudentReportService(mockBackend, mockPDFGen, mockCache, logger)

	studentID := "12345"
	contentHash := "abcd1234"

	// Setup mocks
	mockCache.On("Get", studentID, contentHash).Return(nil, false)

	// Execute
	result := service.tryGetFromCache(studentID, contentHash)

	// Assert
	assert.Nil(t, result)

	// Verify mock expectations
	mockCache.AssertExpectations(t)
}

func TestTryGetFromCache_NilCache(t *testing.T) {
	// Setup
	logger := zap.NewNop()
	mockBackend := new(MockBackendService)
	mockPDFGen := new(MockPDFGenerator)

	service := NewStudentReportService(mockBackend, mockPDFGen, nil, logger)

	// Execute
	result := service.tryGetFromCache("12345", "abcd1234")

	// Assert
	assert.Nil(t, result)
}

func TestGeneratePDF_Success(t *testing.T) {
	// Setup
	logger := zap.NewNop()
	mockBackend := new(MockBackendService)
	mockPDFGen := new(MockPDFGenerator)
	mockCache := new(MockPDFCache)

	service := NewStudentReportService(mockBackend, mockPDFGen, mockCache, logger)

	student := createTestStudent()
	expectedPDF := []byte("generated pdf content")

	// Setup mocks
	mockPDFGen.On("GenerateStudentReport", student).Return(expectedPDF, nil)

	// Execute
	pdfData, err := service.generatePDF(student)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, expectedPDF, pdfData)

	// Verify mock expectations
	mockPDFGen.AssertExpectations(t)
}

func TestGeneratePDF_Error(t *testing.T) {
	// Setup
	logger := zap.NewNop()
	mockBackend := new(MockBackendService)
	mockPDFGen := new(MockPDFGenerator)
	mockCache := new(MockPDFCache)

	service := NewStudentReportService(mockBackend, mockPDFGen, mockCache, logger)

	student := createTestStudent()
	pdfErr := errors.New("pdf generation failed")

	// Setup mocks
	mockPDFGen.On("GenerateStudentReport", student).Return(nil, pdfErr)

	// Execute
	pdfData, err := service.generatePDF(student)

	// Assert
	require.Error(t, err)
	assert.IsType(t, &serviceErrors.PDFGenerationError{}, err)
	assert.Nil(t, pdfData)

	// Verify mock expectations
	mockPDFGen.AssertExpectations(t)
}

func TestStorePDFInCache_Success(t *testing.T) {
	// Setup
	logger := zap.NewNop()
	mockBackend := new(MockBackendService)
	mockPDFGen := new(MockPDFGenerator)
	mockCache := new(MockPDFCache)

	service := NewStudentReportService(mockBackend, mockPDFGen, mockCache, logger)

	studentID := "12345"
	contentHash := "abcd1234"
	pdfData := []byte("pdf content")

	// Setup mocks
	mockCache.On("Set", studentID, pdfData, contentHash).Return(nil)

	// Execute
	service.storePDFInCache(studentID, contentHash, pdfData)

	// Verify mock expectations
	mockCache.AssertExpectations(t)
}

func TestStorePDFInCache_Error(t *testing.T) {
	// Setup
	logger := zap.NewNop()
	mockBackend := new(MockBackendService)
	mockPDFGen := new(MockPDFGenerator)
	mockCache := new(MockPDFCache)

	service := NewStudentReportService(mockBackend, mockPDFGen, mockCache, logger)

	studentID := "12345"
	contentHash := "abcd1234"
	pdfData := []byte("pdf content")
	cacheErr := errors.New("cache write failed")

	// Setup mocks
	mockCache.On("Set", studentID, pdfData, contentHash).Return(cacheErr)

	// Execute - should not panic
	service.storePDFInCache(studentID, contentHash, pdfData)

	// Verify mock expectations
	mockCache.AssertExpectations(t)
}

func TestStorePDFInCache_NilCache(t *testing.T) {
	// Setup
	logger := zap.NewNop()
	mockBackend := new(MockBackendService)
	mockPDFGen := new(MockPDFGenerator)

	service := NewStudentReportService(mockBackend, mockPDFGen, nil, logger)

	// Execute - should not panic
	service.storePDFInCache("12345", "abcd1234", []byte("pdf content"))
}

func TestBuildFileName(t *testing.T) {
	// Setup
	logger := zap.NewNop()
	mockBackend := new(MockBackendService)
	mockPDFGen := new(MockPDFGenerator)
	mockCache := new(MockPDFCache)

	service := NewStudentReportService(mockBackend, mockPDFGen, mockCache, logger)

	testCases := []struct {
		name      string
		studentID string
		expected  string
	}{
		{
			name:      "numeric student ID",
			studentID: "12345",
			expected:  "student_12345_report.pdf",
		},
		{
			name:      "alphanumeric student ID",
			studentID: "STU123",
			expected:  "student_STU123_report.pdf",
		},
		{
			name:      "student ID with special chars",
			studentID: "2024-001",
			expected:  "student_2024-001_report.pdf",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fileName := service.buildFileName(tc.studentID)
			assert.Equal(t, tc.expected, fileName)
		})
	}
}
