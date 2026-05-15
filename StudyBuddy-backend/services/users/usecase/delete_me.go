package usecase

import "context"

// DeleteMe deletes (or deactivates) the current user's account/profile.
// Exact semantics are implemented by the repository (e.g. set is_active=false).
type DeleteMe interface {
	DeleteMe(ctx context.Context, userID string) error
}

type deleteMe struct {
	repo ProfileRepository
}

// NewDeleteMe creates the DeleteMe use case.
func NewDeleteMe(repo ProfileRepository) DeleteMe {
	return &deleteMe{repo: repo}
}

func (u *deleteMe) DeleteMe(ctx context.Context, userID string) error {
	return u.repo.DeleteByUserID(ctx, userID)
}
