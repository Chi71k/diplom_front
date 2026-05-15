package usecase

import "context"
import "studybuddy/backend/services/courses/domain"

// GetCourse defines the use case for retrieving a single course.
type GetCourse interface {
	Get(ctx context.Context, id string) (*domain.Course, error)
}
