package service

import (
	"github.com/wbentaleb/student-report-service/internal/dto"
)

type PDFGenerator interface {
	GenerateStudentReport(student *dto.Student) ([]byte, error)
}
