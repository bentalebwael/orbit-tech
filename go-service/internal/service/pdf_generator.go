package service

import (
	"bytes"
	"fmt"
	"time"

	"github.com/jung-kurt/gofpdf"
	"go.uber.org/zap"

	"github.com/wbentaleb/student-report-service/internal/dto"
)

type PDFService struct {
	logger *zap.Logger
}

func NewPDFService(logger *zap.Logger) *PDFService {
	return &PDFService{
		logger: logger,
	}
}

func (s *PDFService) GenerateStudentReport(student *dto.Student) ([]byte, error) {
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
	s.addTableRow(pdf, "Full Name", s.formatValue(student.Name))
	s.addTableRow(pdf, "Email", s.formatValue(student.Email))
	s.addTableRow(pdf, "Date of Birth", s.formatDate(student.DOB))
	s.addTableRow(pdf, "Gender", s.formatValue(student.Gender))
	s.addTableRow(pdf, "Phone", s.formatValue(student.Phone))
	s.addTableRow(pdf, "System Access", s.formatBool(student.SystemAccess))
	pdf.Ln(5)

	// Academic Information Section
	s.addSectionHeader(pdf, "Academic Information")
	s.addTableRow(pdf, "Class", s.formatValue(student.Class))
	s.addTableRow(pdf, "Section", s.formatValue(student.Section))
	s.addTableRow(pdf, "Roll Number", s.formatIntValue(student.Roll))
	s.addTableRow(pdf, "Admission Date", s.formatDate(student.AdmissionDate))
	s.addTableRow(pdf, "Added By", s.formatValue(student.ReporterName))
	pdf.Ln(5)

	// Parent Information Section
	s.addSectionHeader(pdf, "Parent Information")
	s.addTableRow(pdf, "Father's Name", s.formatValue(student.FatherName))
	s.addTableRow(pdf, "Father's Phone", s.formatValue(student.FatherPhone))
	s.addTableRow(pdf, "Mother's Name", s.formatValue(student.MotherName))
	s.addTableRow(pdf, "Mother's Phone", s.formatValue(student.MotherPhone))
	pdf.Ln(5)

	// Guardian Information Section
	s.addSectionHeader(pdf, "Guardian Information")
	s.addTableRow(pdf, "Guardian Name", s.formatValue(student.GuardianName))
	s.addTableRow(pdf, "Guardian Phone", s.formatValue(student.GuardianPhone))
	s.addTableRow(pdf, "Relationship", s.formatValue(student.RelationOfGuardian))
	pdf.Ln(5)

	// Address Information Section
	s.addSectionHeader(pdf, "Address Information")
	s.addTableRow(pdf, "Current Address", s.formatValue(student.CurrentAddress))
	s.addTableRow(pdf, "Permanent Address", s.formatValue(student.PermanentAddress))

	// Footer - positioned at bottom of current page
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

func (s *PDFService) formatDate(isoDate string) string {
	if isoDate == "" {
		return "N/A"
	}
	t, err := time.Parse(time.RFC3339, isoDate)
	if err != nil {
		s.logger.Warn("Failed to parse date", zap.String("date", isoDate), zap.Error(err))
		return isoDate
	}
	return t.Format("January 2, 2006")
}

func (s *PDFService) formatValue(value string) string {
	if value == "" {
		return "N/A"
	}
	return value
}

func (s *PDFService) formatIntValue(value int) string {
	if value == 0 {
		return "N/A"
	}
	return fmt.Sprintf("%d", value)
}

func (s *PDFService) formatBool(value bool) string {
	if value {
		return "Active"
	}
	return "Inactive"
}

func (s *PDFService) addSectionHeader(pdf *gofpdf.Fpdf, title string) {
	pdf.SetFont("Arial", "B", 14)
	pdf.SetFillColor(52, 152, 219)
	pdf.SetTextColor(255, 255, 255)
	pdf.CellFormat(190, 8, title, "1", 1, "L", true, 0, "")
	pdf.SetTextColor(0, 0, 0)
}

func (s *PDFService) addTableRow(pdf *gofpdf.Fpdf, label, value string) {
	pdf.SetFont("Arial", "B", 11)
	pdf.SetFillColor(236, 240, 241)
	pdf.CellFormat(60, 8, label, "1", 0, "L", true, 0, "")

	pdf.SetFont("Arial", "", 11)
	pdf.SetFillColor(255, 255, 255)
	pdf.CellFormat(130, 8, value, "1", 1, "L", true, 0, "")
}
