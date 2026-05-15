package domain

import "errors"

var (
	ErrFriendshipNotFound  = errors.New("Friendship not found")
	ErrMatchNotFound       = errors.New("Match not found")
	ErrForbidden           = errors.New("Forbidden")
	ErrMatchAlreadyExists  = errors.New("A pending or accepted match already exists between these users")
	ErrCannotMatchSelf     = errors.New("Cannot send a match request to yourself")
	ErrInvalidStatusChange = errors.New("Invalid status transition")
)
