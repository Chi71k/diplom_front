package domain

import (
	"errors"
	"time"
)

type Transaction struct {
	ID        string
	UserID    string
	Amount    int
	Reason    Reason
	SourceKey string
	CreatedAt time.Time
}

type Reason string

const (
	ReasonReviewLeft       Reason = "review_left"
	ReasonReviewReceived   Reason = "review_received"
	ReasonGroupCreated     Reason = "group_created"
	ReasonGroupActivity    Reason = "group_activity"
	ReasonSessionConfirmed Reason = "session_confirmed"
	ReasonMatchAccepted    Reason = "match_accepted"
)

var ErrInvalidPointsRequest = errors.New("invalid points request")

// LeaderboardEntry is one row from the materialized leaderboard view.
type LeaderboardEntry struct {
	UserID      string `json:"userId"`
	TotalPoints int64  `json:"totalPoints"`
	Rank        int    `json:"rank"`
}
