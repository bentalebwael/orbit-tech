package service

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"github.com/wbentaleb/student-report-service/internal/cache"
	"github.com/wbentaleb/student-report-service/internal/dto"
	"github.com/wbentaleb/student-report-service/internal/errors"
	"github.com/wbentaleb/student-report-service/internal/external"
)

type StudentReportService struct {
	backendClient external.BackendService
	pdfGenerator  PDFGenerator
	pdfCache      cache.PDFCache
	logger        *zap.Logger
}

func NewStudentReportService(
	backendClient external.BackendService,
	pdfGenerator PDFGenerator,
	pdfCache cache.PDFCache,
	logger *zap.Logger,
) *StudentReportService {
	return &StudentReportService{
		backendClient: backendClient,
		pdfGenerator:  pdfGenerator,
		pdfCache:      pdfCache,
		logger:        logger,
	}
}

func (s *StudentReportService) GenerateStudentReport(ctx context.Context, studentID string) (pdfData []byte, fileName string, err error) {
	// fetch student data from backend
	student, err := s.fetchStudentData(ctx, studentID)
	if err != nil {
		return nil, "", err
	}

	contentHash := cache.GenerateStudentHash(student)

	// try to retrieve from cache
	if cachedPDF := s.tryGetFromCache(studentID, contentHash); cachedPDF != nil {
		s.logger.Info("Report served from cache",
			zap.String("student_id", studentID),
			zap.String("content_hash", contentHash))

		return cachedPDF, s.buildFileName(studentID), nil
	}

	// if no cache found, generate new PDF
	pdfData, err = s.generatePDF(student)
	if err != nil {
		return nil, "", err
	}

	// store in cache (non-blocking, failure is acceptable)
	s.storePDFInCache(studentID, contentHash, pdfData)

	s.logger.Info("Report generated successfully",
		zap.String("student_id", studentID),
		zap.Int("pdf_size_bytes", len(pdfData)))

	return pdfData, s.buildFileName(studentID), nil
}

func (s *StudentReportService) fetchStudentData(ctx context.Context, studentID string) (*dto.Student, error) {
	student, err := s.backendClient.GetStudent(ctx, studentID)
	if err != nil {
		s.logger.Error("Failed to fetch student data",
			zap.String("student_id", studentID),
			zap.Error(err))
		return nil, fmt.Errorf("failed to fetch student data: %w", err)
	}
	return student, nil
}

func (s *StudentReportService) tryGetFromCache(studentID, contentHash string) []byte {
	if s.pdfCache == nil {
		return nil
	}

	pdfData, found := s.pdfCache.Get(studentID, contentHash)
	if !found {
		s.logger.Debug("Cache miss",
			zap.String("student_id", studentID),
			zap.String("content_hash", contentHash))
		return nil
	}

	return pdfData
}

func (s *StudentReportService) generatePDF(student *dto.Student) ([]byte, error) {
	pdfData, err := s.pdfGenerator.GenerateStudentReport(student)
	if err != nil {
		s.logger.Error("PDF generation failed",
			zap.Int("student_id", student.ID),
			zap.Error(err))
		return nil, errors.NewPDFGenerationError(err)
	}
	return pdfData, nil
}

func (s *StudentReportService) storePDFInCache(studentID, contentHash string, pdfData []byte) {
	if s.pdfCache == nil {
		return
	}

	if err := s.pdfCache.Set(studentID, pdfData, contentHash); err != nil {
		s.logger.Warn("Failed to cache PDF (non-critical)",
			zap.String("student_id", studentID),
			zap.Error(err))
		// Continue - caching failure should not break the flow
	}
}

func (s *StudentReportService) buildFileName(studentID string) string {
	return fmt.Sprintf("student_%s_report.pdf", studentID)
}
