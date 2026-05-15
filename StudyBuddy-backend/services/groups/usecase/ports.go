package usecase

import (
	"context"

	"studybuddy/backend/services/groups/domain"
)

// GroupRepository persists study groups and membership.
type GroupRepository interface {
	Create(ctx context.Context, name, description, ownerID string, courseIDs []string) (*domain.Group, error)
	GetByID(ctx context.Context, id string) (*domain.Group, error)
	ListForUser(ctx context.Context, userID string) ([]domain.Group, error)
	UpdateMetadataAndCourses(ctx context.Context, groupID string, name, description *string, courseIDs *[]string) error
	Delete(ctx context.Context, groupID string) error
	AddMember(ctx context.Context, groupID, userID string, role domain.MemberRole) error
	RemoveMember(ctx context.Context, groupID, userID string) error
	CountMembers(ctx context.Context, groupID string) (int, error)

	ListCourseTitlesForGroup(ctx context.Context, groupID string) ([]string, error)
	ListDistinctInterestNamesForGroup(ctx context.Context, groupID string) ([]string, error)
	ListCandidateUserIDs(ctx context.Context, groupID string, limit int) ([]string, error)
	ListCourseOverlapCandidates(ctx context.Context, groupID string, limit int) ([]OverlapCandidate, error)
	ListProfiles(ctx context.Context, userIDs []string) ([]ProfileSnippet, error)
}

// OverlapCandidate is used when semantic embeddings are unavailable.
type OverlapCandidate struct {
	UserID       string
	OverlapCount int
}

// ProfileSnippet is minimal user info for API responses.
type ProfileSnippet struct {
	UserID    string
	FirstName string
	LastName  string
	AvatarURL string
}

// EmbeddingProvider loads user embedding vectors.
type EmbeddingProvider interface {
	GetOrCompute(ctx context.Context, userID string) ([]float64, error)
}
