package usecase

import (
	"context"
	"fmt"
	"studybuddy/backend/services/matching/domain"
)

type RespondToMatchInput struct {
	MatchID     string
	ResponderID string
	Accept      bool // true = accept, false = decline
}

type RespondToMatch interface {
	Respond(ctx context.Context, in RespondToMatchInput) (*domain.Match, error)
}

type respondToMatch struct {
	repo     MatchRepository
	accepter MatchAccepter
}

func NewRespondToMatch(repo MatchRepository, accepter MatchAccepter) RespondToMatch {
	return &respondToMatch{repo: repo, accepter: accepter}
}

func (uc *respondToMatch) Respond(ctx context.Context, in RespondToMatchInput) (*domain.Match, error) {
	m, err := uc.repo.GetByID(ctx, in.MatchID)
	if err != nil {
		return nil, fmt.Errorf("get match: %w", err)
	}
	if m == nil {
		return nil, domain.ErrMatchNotFound
	}
	// Only the receiver may respond.
	if m.ReceiverID != in.ResponderID {
		return nil, domain.ErrForbidden
	}
	if m.Status != domain.MatchStatusPending {
		return nil, domain.ErrInvalidStatusChange
	}

	if in.Accept {
		if err := uc.accepter.AcceptAndBefriend(ctx, m.ID, m.RequesterID, m.ReceiverID); err != nil {
			return nil, fmt.Errorf("accept match: %w", err)
		}
		m.Status = domain.MatchStatusAccepted
		return m, nil
	}

	if err := uc.repo.UpdateStatus(ctx, m.ID, domain.MatchStatusDeclined); err != nil {
		return nil, fmt.Errorf("update match status: %w", err)
	}
	m.Status = domain.MatchStatusDeclined
	return m, nil
}
