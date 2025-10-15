package service

import (
	"context"

	"github.com/wbentaleb/student-report-service/internal/dto"
)

type PDFGenerator interface {
	GenerateStudentReport(student *dto.Student) ([]byte, error)
}

type ReportService interface {
	GenerateStudentReport(ctx context.Context, studentID string) (pdfData []byte, fileName string, err error)
}
