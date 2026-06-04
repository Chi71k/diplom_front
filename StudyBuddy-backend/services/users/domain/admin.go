package domain

import (
	"errors"
	"time"
)

// ErrUserNotFound is returned when an admin operation targets a missing user.
var ErrUserNotFound = errors.New("user not found")

// AdminUserFilter optionally restricts admin user listing.
type AdminUserFilter struct {
	Role     *string
	IsActive *bool
}

// UserSummary is a row returned by admin user listing.
type UserSummary struct {
	ID        string
	Email     string
	FirstName string
	LastName  string
	Role      string
	IsActive  bool
	CreatedAt time.Time
}

// PlatformStats holds aggregate counts for the admin dashboard.
type PlatformStats struct {
	TotalUsers    int
	ActiveUsers   int
	TotalMatches  int
	TotalGroups   int
	TotalSessions int
}
