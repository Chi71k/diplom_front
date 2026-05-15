package usecase

import "context"
import "studybuddy/backend/services/courses/domain"

// CourseRepository is the port for course persistence.
type CourseRepository interface {
	Create(ctx context.Context, course *domain.Course) error
	Update(ctx context.Context, course *domain.Course) error
	Delete(ctx context.Context, id string) error
	GetByID(ctx context.Context, id string) (*domain.Course, error)
	List(ctx context.Context, filter ListCoursesFilter) ([]domain.Course, error)
}

// ListCoursesFilter defines basic filters for listing courses.
type ListCoursesFilter struct {
	Subject string
	Level   string
	Limit   int
	Offset  int
}
