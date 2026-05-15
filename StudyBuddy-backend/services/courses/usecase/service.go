package usecase

import (
	"context"
	"log"

	"studybuddy/backend/pkg/embedding"
	"studybuddy/backend/services/courses/domain"
)

// Service aggregates all course use cases.
type Service struct {
	repo  CourseRepository
	cache embedding.Cache
}

// NewService creates a new Service.
func NewService(repo CourseRepository, cache embedding.Cache) *Service {
	return &Service{repo: repo, cache: cache}
}

// Ensure Service implements interfaces.
var (
	_ CreateCourse = (*Service)(nil)
	_ GetCourse    = (*Service)(nil)
	_ ListCourses  = (*Service)(nil)
	_ UpdateCourse = (*Service)(nil)
	_ DeleteCourse = (*Service)(nil)
)

func (s *Service) Create(ctx context.Context, input CreateCourseInput) (*domain.Course, error) {
	c := &domain.Course{
		Title:       input.Title,
		Description: input.Description,
		Subject:     input.Subject,
		Level:       input.Level,
		OwnerUserID: input.OwnerUserID,
	}
	if err := s.repo.Create(ctx, c); err != nil {
		return nil, err
	}
	if s.cache != nil {
		if err := s.cache.Delete(ctx, c.OwnerUserID); err != nil {
			log.Printf("embedding cache invalidate after create course: %v", err)
		}
	}
	return c, nil
}

func (s *Service) Get(ctx context.Context, id string) (*domain.Course, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) List(ctx context.Context, filter ListCoursesFilter) ([]domain.Course, error) {
	return s.repo.List(ctx, filter)
}

func (s *Service) Update(ctx context.Context, input UpdateCourseInput) (*domain.Course, error) {
	existing, err := s.repo.GetByID(ctx, input.ID)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, domain.ErrCourseNotFound
	}
	if existing.OwnerUserID != input.RequestingUser {
		return nil, domain.ErrForbidden
	}
	if input.Title != nil {
		existing.Title = *input.Title
	}
	if input.Description != nil {
		existing.Description = *input.Description
	}
	if input.Subject != nil {
		existing.Subject = *input.Subject
	}
	if input.Level != nil {
		existing.Level = *input.Level
	}
	if err := s.repo.Update(ctx, existing); err != nil {
		return nil, err
	}
	if s.cache != nil {
		if err := s.cache.Delete(ctx, existing.OwnerUserID); err != nil {
			log.Printf("embedding cache invalidate after update course: %v", err)
		}
	}
	return existing, nil
}

func (s *Service) Delete(ctx context.Context, input DeleteCourseInput) error {
	existing, err := s.repo.GetByID(ctx, input.ID)
	if err != nil {
		return err
	}
	if existing == nil {
		return domain.ErrCourseNotFound
	}
	if existing.OwnerUserID != input.RequestingUser {
		return domain.ErrForbidden
	}
	if err := s.repo.Delete(ctx, input.ID); err != nil {
		return err
	}
	if s.cache != nil {
		if err := s.cache.Delete(ctx, existing.OwnerUserID); err != nil {
			log.Printf("embedding cache invalidate after delete course: %v", err)
		}
	}
	return nil
}
