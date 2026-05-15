package usecase

import (
	"context"

	"studybuddy/backend/services/matching/domain"
)

type MatchRepository interface {
	Create(ctx context.Context, m *domain.Match) error
	GetByID(ctx context.Context, id string) (*domain.Match, error)
	// GetBetween is used for pending/accepted match between the two users
	GetBetween(ctx context.Context, userA, userB string) (*domain.Match, error)
	UpdateStatus(ctx context.Context, id string, status domain.MatchStatus) error
	ListForUser(ctx context.Context, userID string, filter ListMatchesFilter) ([]domain.Match, error)
}

// MatchAccepter updates a pending match to accepted and creates bidirectional friendships atomically.
type MatchAccepter interface {
	AcceptAndBefriend(ctx context.Context, matchID, requesterID, receiverID string) error
}

// ListMatchesFilter narrows ListForUser results
type ListMatchesFilter struct {
	// Status filters by a specific status. Empty string means "all"
	Status domain.MatchStatus
	Limit  int
	Offset int
}

// ProfileClient fetches the lightweight profile data from the Users service
type ProfileClient interface {
	GetProfile(ctx context.Context, userID string) (*ProfileData, error)
	GetInterestIDs(ctx context.Context, userID string) ([]string, error)
	ListInterestNamesForUser(ctx context.Context, userID string) ([]string, error)
}

// ProfileData is the minimal profile needed for candidate enrichment
type ProfileData struct {
	UserID    string
	FirstName string
	LastName  string
	Bio       string
	AvatarURL string
}

// SlotClient fetches the availability slots from the Availability service
type SlotClient interface {
	ListForUser(ctx context.Context, userID string) ([]SlotData, error)
	ListForUsers(ctx context.Context, userIDs []string) ([]SlotData, error)
}

// SlotData is the minimal slot needed for overlap scoring
type SlotData struct {
	ID        string
	UserID    string
	DayOfWeek int
	StartTime string
	EndTime   string
	Timezone  string
}

// CourseClient fetches courses the user owns/enrolled in
// For MVP, a user "has" courses they created
type CourseClient interface {
	ListCourseIDsForUser(ctx context.Context, userID string) ([]string, error)
	ListCourseTitlesForUser(ctx context.Context, userID string) ([]string, error)
}

// CandidateStore provides a pre-filtered pool of candidate user IDs
// Lists all active users except the requester and those already matched
type CandidateStore interface {
	ListCandidateIDs(ctx context.Context, requesterID string, excludeIDs []string) ([]string, error)
}

// EmbeddingProvider computes or retrieves a cached semantic embedding for a user.
// Returns nil embedding without error when embedding is unavailable (graceful degradation).
type EmbeddingProvider interface {
	GetOrCompute(ctx context.Context, userID string) ([]float64, error)
}

// ReputationClient returns a normalized reputation score in [0, 1] for a user (e.g. average rating).
type ReputationClient interface {
	GetAverageRating(ctx context.Context, userID string) float64
}

// FriendshipRepository tracks mutual connections between users.
type FriendshipRepository interface {
	CreateBoth(ctx context.Context, userA, userB string) error
	ListFriends(ctx context.Context, userID string) ([]string, error)
	Delete(ctx context.Context, userID, friendID string) error
	MutualFriendCount(ctx context.Context, userA, userB string) (int, error)
}
