package domain

import "time"

type MatchStatus string

const (
	MatchStatusPending  MatchStatus = "pending"
	MatchStatusAccepted MatchStatus = "accepted"
	MatchStatusDeclined MatchStatus = "declined"
	MatchStatusCanceled MatchStatus = "canceled"
)

type Match struct {
	ID          string
	RequesterID string
	ReceiverID  string
	Status      MatchStatus
	Message     string // optional invitation message
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type MatchCandidate struct {
	UserID             string
	FirstName          string
	LastName           string
	Bio                string
	AvatarURL          string
	CommonCourses      []string
	CommonSlots        []SlotOverlap
	SemanticScore      float64 // 0-1, cosine similarity of stored embeddings
	AvailScore         float64 // 0-1, temporal overlap ratio
	CourseScore        float64 // 0-1, shared courses proportion (Jaccard)
	ReputationScore    float64 // 0-1, normalised peer rating
	MutualFriendsScore float64 // 0-1, normalised mutual friend count
	OverallScore       float64 // weighted composite
}

type SlotOverlap struct {
	DayOfWeek int // 0=Monday ... 6=Sunday
	StartTime string
	EndTime   string
	Timezone  string
}
