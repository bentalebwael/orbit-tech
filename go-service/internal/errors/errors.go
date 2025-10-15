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

// IsNotFound checks if the error is or wraps a NotFoundError
func IsNotFound(err error) bool {
	var notFoundErr *NotFoundError
	return errors.As(err, &notFoundErr)
}

// IsServiceError checks if the error is or wraps a ServiceError
func IsServiceError(err error) bool {
	var serviceErr *ServiceError
	return errors.As(err, &serviceErr)
}
