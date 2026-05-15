package usecase

import (
	"context"
	"errors"
	"log"
	"sort"
	"strings"

	"studybuddy/backend/pkg/embedding"
	"studybuddy/backend/services/users/domain"
)

var ErrInvalidInterestIDs = errors.New("invalid interest ids")

type ListInterests interface {
	ListInterests(ctx context.Context) ([]domain.Interest, error)
}

type listInterests struct {
	repo InterestRepository
}

func NewListInterests(repo InterestRepository) ListInterests {
	return &listInterests{repo: repo}
}

func (l *listInterests) ListInterests(ctx context.Context) ([]domain.Interest, error) {
	return l.repo.ListAll(ctx)
}

type GetMyInterests interface {
	GetMyInterests(ctx context.Context, userID string) ([]domain.Interest, error)
}

type getMyInterests struct {
	repo UserInterestRepository
}

func NewGetMyInterests(repo UserInterestRepository) GetMyInterests {
	return &getMyInterests{repo: repo}
}

func (g *getMyInterests) GetMyInterests(ctx context.Context, userID string) ([]domain.Interest, error) {
	return g.repo.ListForUser(ctx, userID)
}

type ReplaceMyInterestsInput struct {
	UserID      string
	InterestIDs []string
}

type ReplaceMyInterests interface {
	ReplaceMyInterests(ctx context.Context, in ReplaceMyInterestsInput) ([]domain.Interest, error)
}

type replaceMyInterests struct {
	interestRepo InterestRepository
	userRepo     UserInterestRepository
	cache        embedding.Cache
}

func NewReplaceMyInterests(interestRepo InterestRepository, userRepo UserInterestRepository, cache embedding.Cache) ReplaceMyInterests {
	return &replaceMyInterests{interestRepo: interestRepo, userRepo: userRepo, cache: cache}
}

func (u *replaceMyInterests) ReplaceMyInterests(ctx context.Context, in ReplaceMyInterestsInput) ([]domain.Interest, error) {
	ids := normalizeIDs(in.InterestIDs)

	existing, err := u.interestRepo.GetByIDs(ctx, ids)
	if err != nil {
		return nil, err
	}
	if len(existing) != len(ids) {
		return nil, ErrInvalidInterestIDs
	}

	if err := u.userRepo.ReplaceForUser(ctx, in.UserID, ids); err != nil {
		return nil, err
	}

	if u.cache != nil {
		if err := u.cache.Delete(ctx, in.UserID); err != nil {
			log.Printf("embedding cache invalidate after interests replace: %v", err)
		}
	}

	sort.Slice(existing, func(i, j int) bool { return existing[i].Name < existing[j].Name })
	return existing, nil
}

func normalizeIDs(ids []string) []string {
	seen := make(map[string]struct{}, len(ids))
	out := make([]string, 0, len(ids))
	for _, id := range ids {
		id = strings.TrimSpace(id)
		if id == "" {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	return out
}
