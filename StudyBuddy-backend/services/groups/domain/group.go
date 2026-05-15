package domain

import "time"

type Group struct {
	ID          string
	Name        string
	Description string
	OwnerID     string
	CourseIDs   []string
	Members     []Member
	CreatedAt   time.Time
}

type Member struct {
	UserID   string
	Role     MemberRole
	JoinedAt time.Time
}

type MemberRole string

const (
	RoleOwner  MemberRole = "owner"
	RoleMember MemberRole = "member"
)
