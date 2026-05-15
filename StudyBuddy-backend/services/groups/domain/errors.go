package domain

import "errors"

var (
	ErrGroupNotFound    = errors.New("group not found")
	ErrGroupFull        = errors.New("group is full (max 5 members)")
	ErrAlreadyMember    = errors.New("user is already a member")
	ErrNotMember        = errors.New("user is not a member")
	ErrNotOwner         = errors.New("only the group owner can do this")
	ErrMinMembers       = errors.New("group must have at least 2 members")
	ErrForbidden        = errors.New("forbidden")
	ErrOwnerCannotLeave = errors.New("owner cannot leave the group; delete the group instead")
	ErrInvalidGroupName = errors.New("invalid group name")
)
