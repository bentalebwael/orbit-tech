package service

import (
	"bytes"
	"strings"
	"testing"

	"github.com/jung-kurt/gofpdf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/wbentaleb/student-report-service/internal/dto"
)

func TestNewPDFService(t *testing.T) {
	logger := zap.NewNop()
	service := NewPDFService(logger)

	assert.NotNil(t, service)
	assert.Equal(t, logger, service.logger)
}

func TestGenerateStudentReport_Success(t *testing.T) {
	// Setup
	logger := zap.NewNop()
	service := NewPDFService(logger)

	student := &dto.Student{
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
		CurrentAddress:     "123 Main St, City",
		PermanentAddress:   "456 Oak Ave, Town",
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

	// Execute
	pdfData, err := service.GenerateStudentReport(student)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, pdfData)
	assert.Greater(t, len(pdfData), 0)

	// Verify it's a valid PDF by checking the header
	assert.True(t, bytes.HasPrefix(pdfData, []byte("%PDF-")))
}

func TestGenerateStudentReport_MinimalData(t *testing.T) {
	// Setup
	logger := zap.NewNop()
	service := NewPDFService(logger)

	// Minimal student with only required fields
	student := &dto.Student{
		ID:   99999,
		Name: "Test Student",
	}

	// Execute
	pdfData, err := service.GenerateStudentReport(student)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, pdfData)
	assert.Greater(t, len(pdfData), 0)

	// Verify it's a valid PDF
	assert.True(t, bytes.HasPrefix(pdfData, []byte("%PDF-")))
}

func TestGenerateStudentReport_AllFieldsEmpty(t *testing.T) {
	// Setup
	logger := zap.NewNop()
	service := NewPDFService(logger)

	// Student with all empty/zero values
	student := &dto.Student{}

	// Execute
	pdfData, err := service.GenerateStudentReport(student)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, pdfData)
	assert.Greater(t, len(pdfData), 0)

	// Verify it's a valid PDF
	assert.True(t, bytes.HasPrefix(pdfData, []byte("%PDF-")))
}

func TestFormatDate_Success(t *testing.T) {
	// Setup
	logger := zap.NewNop()
	service := NewPDFService(logger)

	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "valid ISO date",
			input:    "2024-01-15T10:30:00Z",
			expected: "January 15, 2024",
		},
		{
			name:     "valid date with timezone",
			input:    "2023-12-25T00:00:00+05:00",
			expected: "December 25, 2023",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "N/A",
		},
		{
			name:     "invalid date format",
			input:    "2024-01-15",
			expected: "2024-01-15", // Returns original when parsing fails
		},
		{
			name:     "invalid date string",
			input:    "invalid-date",
			expected: "invalid-date", // Returns original when parsing fails
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := service.formatDate(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestFormatValue(t *testing.T) {
	// Setup
	logger := zap.NewNop()
	service := NewPDFService(logger)

	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "normal string",
			input:    "John Doe",
			expected: "John Doe",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "N/A",
		},
		{
			name:     "whitespace only",
			input:    "   ",
			expected: "   ",
		},
		{
			name:     "special characters",
			input:    "Test@123!#$",
			expected: "Test@123!#$",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := service.formatValue(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestFormatIntValue(t *testing.T) {
	// Setup
	logger := zap.NewNop()
	service := NewPDFService(logger)

	testCases := []struct {
		name     string
		input    int
		expected string
	}{
		{
			name:     "positive number",
			input:    42,
			expected: "42",
		},
		{
			name:     "zero",
			input:    0,
			expected: "N/A",
		},
		{
			name:     "negative number",
			input:    -5,
			expected: "-5",
		},
		{
			name:     "large number",
			input:    999999,
			expected: "999999",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := service.formatIntValue(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestFormatBool(t *testing.T) {
	// Setup
	logger := zap.NewNop()
	service := NewPDFService(logger)

	testCases := []struct {
		name     string
		input    bool
		expected string
	}{
		{
			name:     "true value",
			input:    true,
			expected: "Active",
		},
		{
			name:     "false value",
			input:    false,
			expected: "Inactive",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := service.formatBool(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestAddSectionHeader(t *testing.T) {
	// Setup
	logger := zap.NewNop()
	service := NewPDFService(logger)

	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	// Execute
	service.addSectionHeader(pdf, "Test Section")

	// Get PDF output
	var buf bytes.Buffer
	err := pdf.Output(&buf)

	// Assert
	require.NoError(t, err)
	assert.Greater(t, buf.Len(), 0)

	// Verify it's a valid PDF
	assert.True(t, bytes.HasPrefix(buf.Bytes(), []byte("%PDF-")))
}

func TestAddTableRow(t *testing.T) {
	// Setup
	logger := zap.NewNop()
	service := NewPDFService(logger)

	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	// Execute
	service.addTableRow(pdf, "Test Label", "Test Value")

	// Get PDF output
	var buf bytes.Buffer
	err := pdf.Output(&buf)

	// Assert
	require.NoError(t, err)
	assert.Greater(t, buf.Len(), 0)

	// Verify it's a valid PDF
	assert.True(t, bytes.HasPrefix(buf.Bytes(), []byte("%PDF-")))
}

func TestGenerateStudentReport_LongValues(t *testing.T) {
	// Setup
	logger := zap.NewNop()
	service := NewPDFService(logger)

	// Create student with very long values
	longString := strings.Repeat("A very long string that goes on and on ", 10)
	student := &dto.Student{
		ID:                 12345,
		Name:               longString,
		Email:              "verylongemailaddress@extremelylongdomainname.com",
		CurrentAddress:     longString,
		PermanentAddress:   longString,
		GuardianName:       longString,
		RelationOfGuardian: longString,
	}

	// Execute
	pdfData, err := service.GenerateStudentReport(student)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, pdfData)
	assert.Greater(t, len(pdfData), 0)

	// Verify it's a valid PDF
	assert.True(t, bytes.HasPrefix(pdfData, []byte("%PDF-")))
}

func TestGenerateStudentReport_SpecialCharacters(t *testing.T) {
	// Setup
	logger := zap.NewNop()
	service := NewPDFService(logger)

	// Create student with special characters
	student := &dto.Student{
		ID:               12345,
		Name:             "José García-López",
		Email:            "test@español.com",
		CurrentAddress:   "123 Ñoño Street, São Paulo",
		FatherName:       "François Müller",
		MotherName:       "Anna Żółć",
		GuardianName:     "李明 (Li Ming)",
		Phone:            "+1-234-567-8900",
		Gender:           "Male/Masculine",
		Class:            "10-A",
		Section:          "A/B",
		ReporterName:     "Admin@2024",
		PermanentAddress: "456 Oak & Elm Street, City #5",
	}

	// Execute
	pdfData, err := service.GenerateStudentReport(student)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, pdfData)
	assert.Greater(t, len(pdfData), 0)

	// Verify it's a valid PDF
	assert.True(t, bytes.HasPrefix(pdfData, []byte("%PDF-")))
}

func TestGenerateStudentReport_FutureDates(t *testing.T) {
	// Setup
	logger := zap.NewNop()
	service := NewPDFService(logger)

	// Create student with future dates
	student := &dto.Student{
		ID:            12345,
		Name:          "Future Student",
		DOB:           "2030-01-01T00:00:00Z",
		AdmissionDate: "2025-09-01T00:00:00Z",
		LastUpdated:   "2025-12-31T23:59:59Z",
	}

	// Execute
	pdfData, err := service.GenerateStudentReport(student)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, pdfData)
	assert.Greater(t, len(pdfData), 0)

	// Verify it's a valid PDF
	assert.True(t, bytes.HasPrefix(pdfData, []byte("%PDF-")))
}

func TestGenerateStudentReport_InvalidDateFormats(t *testing.T) {
	// Setup
	logger := zap.NewNop()
	service := NewPDFService(logger)

	// Create student with invalid date formats
	student := &dto.Student{
		ID:            12345,
		Name:          "Test Student",
		DOB:           "01/01/2000", // Invalid format
		AdmissionDate: "2015/06/01", // Invalid format
		LastUpdated:   "yesterday",  // Invalid format
	}

	// Execute - should handle gracefully
	pdfData, err := service.GenerateStudentReport(student)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, pdfData)
	assert.Greater(t, len(pdfData), 0)

	// Verify it's a valid PDF
	assert.True(t, bytes.HasPrefix(pdfData, []byte("%PDF-")))
}

func TestGenerateStudentReport_ZeroID(t *testing.T) {
	// Setup
	logger := zap.NewNop()
	service := NewPDFService(logger)

	// Create student with zero ID
	student := &dto.Student{
		ID:   0,
		Name: "Zero ID Student",
	}

	// Execute
	pdfData, err := service.GenerateStudentReport(student)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, pdfData)
	assert.Greater(t, len(pdfData), 0)

	// Verify it's a valid PDF
	assert.True(t, bytes.HasPrefix(pdfData, []byte("%PDF-")))
}

func TestGenerateStudentReport_NegativeValues(t *testing.T) {
	// Setup
	logger := zap.NewNop()
	service := NewPDFService(logger)

	// Create student with negative roll number
	student := &dto.Student{
		ID:   12345,
		Name: "Test Student",
		Roll: -10,
	}

	// Execute
	pdfData, err := service.GenerateStudentReport(student)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, pdfData)
	assert.Greater(t, len(pdfData), 0)

	// Verify it's a valid PDF
	assert.True(t, bytes.HasPrefix(pdfData, []byte("%PDF-")))
}

// Benchmark tests
func BenchmarkGenerateStudentReport(b *testing.B) {
	logger := zap.NewNop()
	service := NewPDFService(logger)

	student := &dto.Student{
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

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = service.GenerateStudentReport(student)
	}
}

func BenchmarkFormatDate(b *testing.B) {
	logger := zap.NewNop()
	service := NewPDFService(logger)

	dateStr := "2024-01-15T10:30:00Z"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = service.formatDate(dateStr)
	}
}
