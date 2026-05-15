package usecase

import (
	"context"
	"log"

	"studybuddy/backend/pkg/embedding"
	"studybuddy/backend/services/users/domain"
)

// UpdateMeInput for partial profile update.
type UpdateMeInput struct {
	UserID    string
	FirstName *string
	LastName  *string
	Bio       *string
	AvatarURL *string
}

// UpdateMe updates the profile for the given user ID.
type UpdateMe interface {
	UpdateMe(ctx context.Context, in UpdateMeInput) (*domain.Profile, error)
}

type updateMe struct {
	repo  ProfileRepository
	cache embedding.Cache
}

// NewUpdateMe creates the UpdateMe use case.
func NewUpdateMe(repo ProfileRepository, cache embedding.Cache) UpdateMe {
	return &updateMe{repo: repo, cache: cache}
}

func (u *updateMe) UpdateMe(ctx context.Context, in UpdateMeInput) (*domain.Profile, error) {
	existing, err := u.repo.GetByUserID(ctx, in.UserID)
	if err != nil {
		return nil, err
	}
	profile := &domain.Profile{UserID: in.UserID}
	if existing != nil {
		profile.FirstName = existing.FirstName
		profile.LastName = existing.LastName
		profile.Bio = existing.Bio
		profile.AvatarURL = existing.AvatarURL
		profile.Email = existing.Email
		profile.CreatedAt = existing.CreatedAt
	}
	if in.FirstName != nil {
		profile.FirstName = *in.FirstName
	}
	if in.LastName != nil {
		profile.LastName = *in.LastName
	}
	if in.Bio != nil {
		profile.Bio = *in.Bio
	}
	if in.AvatarURL != nil {
		profile.AvatarURL = *in.AvatarURL
	}
	if err := u.repo.Upsert(ctx, profile); err != nil {
		return nil, err
	}
	out, err := u.repo.GetByUserID(ctx, in.UserID)
	if err != nil {
		return nil, err
	}
	if u.cache != nil {
		if err := u.cache.Delete(ctx, in.UserID); err != nil {
			log.Printf("embedding cache invalidate after profile update: %v", err)
		}
	}
	return out, nil
}
