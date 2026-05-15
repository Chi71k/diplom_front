package domain

import (
	"errors"
	"time"
)

type Review struct {
	ID         string
	ReviewerID string
	RevieweeID string
	MatchID    string
	Rating     int
	Comment    string
	CreatedAt  time.Time
}

type RatingSummary struct {
	UserID        string
	AverageRating float64
	TotalReviews  int
}

var (
	ErrAlreadyReviewed = errors.New("you already reviewed this user for this match")
	ErrNoAcceptedMatch = errors.New("you must have an accepted match with this user to review them")
	ErrInvalidRating   = errors.New("rating must be between 1 and 5")
	ErrCommentTooLong  = errors.New("comment must be at most 500 characters")
	ErrSelfReview      = errors.New("you cannot review yourself")
)
