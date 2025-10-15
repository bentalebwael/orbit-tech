package external

import (
	"context"

	"github.com/wbentaleb/student-report-service/internal/dto"
)

type BackendService interface {
	GetStudent(ctx context.Context, id string) (*dto.Student, error)
	CheckHealth(ctx context.Context) bool
}
