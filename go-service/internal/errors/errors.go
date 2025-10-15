package errors

import (
	"errors"
	"fmt"
)

type NotFoundError struct {
	Resource string
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("%s not found", e.Resource)
}

type ServiceError struct {
	Service string
	Err     error
}

func (e *ServiceError) Error() string {
	return fmt.Sprintf("%s service error: %v", e.Service, e.Err)
}

type PDFGenerationError struct {
	Err error
}

func (e *PDFGenerationError) Error() string {
	return fmt.Sprintf("PDF generation failed: %v", e.Err)
}

func NewPDFGenerationError(err error) *PDFGenerationError {
	return &PDFGenerationError{Err: err}
}

func IsNotFound(err error) bool {
	var notFoundErr *NotFoundError
	return errors.As(err, &notFoundErr)
}

func IsServiceError(err error) bool {
	var serviceErr *ServiceError
	return errors.As(err, &serviceErr)
}

func IsPDFGenerationError(err error) bool {
	var pdfErr *PDFGenerationError
	return errors.As(err, &pdfErr)
}
