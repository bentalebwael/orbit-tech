package service

import (
	"bytes"
	"fmt"
	"time"

	"github.com/jung-kurt/gofpdf"
	"github.com/wbentaleb/student-report-service/internal/domain"
	"go.uber.org/zap"
)

// PDFService handles PDF generation
type PDFService struct {
	logger *zap.Logger
}

// NewPDFService creates a new PDF service instance
func NewPDFService(logger *zap.Logger) *PDFService {
	return &PDFService{
		logger: logger,
	}
}

// GenerateStudentReport generates a PDF report for a student
func (s *PDFService) GenerateStudentReport(student *domain.Student) ([]byte, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	// Set font for header
	pdf.SetFont("Arial", "B", 18)
	pdf.SetTextColor(44, 62, 80)
	pdf.CellFormat(190, 10, "Student Report", "", 1, "C", false, 0, "")
	pdf.Ln(5)

	// Add generation date
	pdf.SetFont("Arial", "", 10)
	pdf.SetTextColor(127, 140, 141)
	currentTime := time.Now().Format("January 2, 2006 at 3:04 PM")
	pdf.CellFormat(190, 6, fmt.Sprintf("Generated on: %s", currentTime), "", 1, "C", false, 0, "")
	pdf.Ln(10)

	// Personal Information Section
	s.addSectionHeader(pdf, "Personal Information")
	s.addTableRow(pdf, "Student ID", fmt.Sprintf("%d", student.ID))
	s.addTableRow(pdf, "Full Name", student.Name)
	s.addTableRow(pdf, "Email", student.Email)
	s.addTableRow(pdf, "Date of Birth", student.DateOfBirth)
	s.addTableRow(pdf, "Gender", student.Gender)
	s.addTableRow(pdf, "Blood Group", student.BloodGroup)
	pdf.Ln(5)

	// Academic Information Section
	s.addSectionHeader(pdf, "Academic Information")
	s.addTableRow(pdf, "Class", student.Class)
	s.addTableRow(pdf, "Section", student.Section)
	s.addTableRow(pdf, "Roll Number", fmt.Sprintf("%d", student.Roll))
	s.addTableRow(pdf, "Admission Date", student.AdmissionDate)
	s.addTableRow(pdf, "Status", student.Status)
	pdf.Ln(5)

	// Guardian Information Section
	s.addSectionHeader(pdf, "Guardian Information")
	s.addTableRow(pdf, "Guardian Name", student.GuardianName)
	s.addTableRow(pdf, "Guardian Phone", student.GuardianPhone)
	s.addTableRow(pdf, "Guardian Email", student.GuardianEmail)
	pdf.Ln(5)

	// Contact Information Section
	s.addSectionHeader(pdf, "Contact Information")
	s.addTableRow(pdf, "Phone", student.Phone)
	s.addTableRow(pdf, "Address", student.Address)
	pdf.Ln(10)

	// Footer
	pdf.SetY(-30)
	pdf.SetFont("Arial", "I", 8)
	pdf.SetTextColor(127, 140, 141)
	pdf.CellFormat(190, 5, "This is an auto-generated report from the Student Management System", "", 1, "C", false, 0, "")
	pdf.CellFormat(190, 5, fmt.Sprintf("Report ID: SR-%d-%d", student.ID, time.Now().Unix()), "", 1, "C", false, 0, "")

	// Generate PDF bytes
	var buf bytes.Buffer
	err := pdf.Output(&buf)
	if err != nil {
		s.logger.Error("Failed to generate PDF", zap.Error(err))
		return nil, fmt.Errorf("failed to generate PDF: %w", err)
	}

	s.logger.Info("PDF generated successfully", zap.Int("student_id", student.ID))
	return buf.Bytes(), nil
}

// addSectionHeader adds a styled section header
func (s *PDFService) addSectionHeader(pdf *gofpdf.Fpdf, title string) {
	pdf.SetFont("Arial", "B", 14)
	pdf.SetFillColor(52, 152, 219)
	pdf.SetTextColor(255, 255, 255)
	pdf.CellFormat(190, 8, title, "1", 1, "L", true, 0, "")
	pdf.SetTextColor(0, 0, 0)
}

// addTableRow adds a two-column table row
func (s *PDFService) addTableRow(pdf *gofpdf.Fpdf, label, value string) {
	pdf.SetFont("Arial", "B", 11)
	pdf.SetFillColor(236, 240, 241)
	pdf.CellFormat(60, 8, label, "1", 0, "L", true, 0, "")

	pdf.SetFont("Arial", "", 11)
	pdf.SetFillColor(255, 255, 255)
	pdf.CellFormat(130, 8, value, "1", 1, "L", true, 0, "")
}
