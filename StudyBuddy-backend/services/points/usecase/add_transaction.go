package usecase

import (
	"context"
	"strings"

	"studybuddy/backend/services/points/domain"
)

var allowedReasons = map[domain.Reason]struct{}{
	domain.ReasonReviewLeft:       {},
	domain.ReasonReviewReceived:   {},
	domain.ReasonGroupCreated:     {},
	domain.ReasonGroupActivity:    {},
	domain.ReasonSessionConfirmed: {},
	domain.ReasonMatchAccepted:    {},
}

type AddTransactionInput struct {
	UserID    string
	Amount    int
	Reason    domain.Reason
	MatchID   string
	SourceKey string
}

type AddTransaction interface {
	AddTransaction(ctx context.Context, in AddTransactionInput) error
}

type addTransaction struct {
	repo PointsRepository
}

func NewAddTransaction(repo PointsRepository) AddTransaction {
	return &addTransaction{repo: repo}
}

func (uc *addTransaction) AddTransaction(ctx context.Context, in AddTransactionInput) error {
	if strings.TrimSpace(in.UserID) == "" || in.Amount == 0 {
		return domain.ErrInvalidPointsRequest
	}
	if _, ok := allowedReasons[in.Reason]; !ok {
		return domain.ErrInvalidPointsRequest
	}
	sk := strings.TrimSpace(in.SourceKey)
	if sk == "" {
		sk = strings.TrimSpace(in.MatchID)
	}
	return uc.repo.AddTransaction(ctx, in.UserID, in.Amount, in.Reason, sk)
}
