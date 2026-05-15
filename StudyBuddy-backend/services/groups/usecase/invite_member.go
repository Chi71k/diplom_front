package usecase

import (
	"context"

	"studybuddy/backend/services/groups/domain"
)

type InviteMemberInput struct {
	ActorID   string
	GroupID   string
	InviteeID string
}

type InviteMember interface {
	InviteMember(ctx context.Context, in InviteMemberInput) error
}

type inviteMember struct {
	repo      GroupRepository
	pointsURL string
	jwtSecret []byte
}

func NewInviteMember(repo GroupRepository, pointsURL string, jwtSecret []byte) InviteMember {
	return &inviteMember{repo: repo, pointsURL: pointsURL, jwtSecret: jwtSecret}
}

func (uc *inviteMember) InviteMember(ctx context.Context, in InviteMemberInput) error {
	g, err := uc.repo.GetByID(ctx, in.GroupID)
	if err != nil {
		return err
	}
	if g == nil {
		return domain.ErrGroupNotFound
	}
	if g.OwnerID != in.ActorID {
		return domain.ErrNotOwner
	}
	if in.InviteeID == in.ActorID {
		return domain.ErrForbidden
	}

	n, err := uc.repo.CountMembers(ctx, in.GroupID)
	if err != nil {
		return err
	}
	if n >= 5 {
		return domain.ErrGroupFull
	}

	err = uc.repo.AddMember(ctx, in.GroupID, in.InviteeID, domain.RoleMember)
	if err != nil {
		return err
	}

	n2, err := uc.repo.CountMembers(ctx, in.GroupID)
	if err != nil {
		return err
	}
	if n2 == 2 {
		FireGroupCreatedPoints(uc.pointsURL, uc.jwtSecret, g.OwnerID, in.GroupID)
	}
	return nil
}
