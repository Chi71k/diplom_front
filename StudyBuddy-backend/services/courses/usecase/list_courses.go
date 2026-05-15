package usecase

import "context"
import "studybuddy/backend/services/courses/domain"

// ListCourses defines the use case for listing courses.
type ListCourses interface {
	List(ctx context.Context, filter ListCoursesFilter) ([]domain.Course, error)
}
