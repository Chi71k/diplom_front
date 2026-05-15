package usecase

import "context"
import "studybuddy/backend/services/courses/domain"

// CreateCourseInput is the input for creating a course.
type CreateCourseInput struct {
	Title       string
	Description string
	Subject     string
	Level       string
	OwnerUserID string
}

// CreateCourse defines the use case for creating a course.
type CreateCourse interface {
	Create(ctx context.Context, input CreateCourseInput) (*domain.Course, error)
}
