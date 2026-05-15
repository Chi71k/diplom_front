package usecase

import (
	"context"
	"log"
	"strings"
	"unicode/utf8"

	"studybuddy/backend/pkg/auth"
	"studybuddy/backend/services/reviews/domain"
)

const maxCommentRunes = 500

type CreateReviewInput struct {
	ReviewerID string
	RevieweeID string
	MatchID    string
	Rating     int
	Comment    string
	AuthHeader string
}

type CreateReview interface {
	CreateReview(ctx context.Context, in CreateReviewInput) (*domain.Review, error)
}

type createReview struct {
	reviews   ReviewRepository
	matches   MatchChecker
	cache     EmbeddingCacheInvalidator
	pointsURL string
	jwtSecret []byte
}

func NewCreateReview(
	reviews ReviewRepository,
	matches MatchChecker,
	cache EmbeddingCacheInvalidator,
	pointsURL string,
	jwtSecret []byte,
) CreateReview {
	return &createReview{
		reviews:   reviews,
		matches:   matches,
		cache:     cache,
		pointsURL: strings.TrimSuffix(pointsURL, "/"),
		jwtSecret: jwtSecret,
	}
}

func (uc *createReview) CreateReview(ctx context.Context, in CreateReviewInput) (*domain.Review, error) {
	if in.ReviewerID == in.RevieweeID {
		return nil, domain.ErrSelfReview
	}
	if in.Rating < 1 || in.Rating > 5 {
		return nil, domain.ErrInvalidRating
	}
	if utf8.RuneCountInString(in.Comment) > maxCommentRunes {
		return nil, domain.ErrCommentTooLong
	}
	ok, err := uc.matches.HasAcceptedMatchBetween(ctx, in.MatchID, in.ReviewerID, in.RevieweeID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, domain.ErrNoAcceptedMatch
	}

	rev, err := uc.reviews.Create(ctx, in.ReviewerID, in.RevieweeID, in.MatchID, in.Rating, in.Comment)
	if err != nil {
		return nil, err
	}

	go uc.sideEffects(in)

	return rev, nil
}

func (uc *createReview) sideEffects(in CreateReviewInput) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultSideEffectTimeout)
	defer cancel()

	if uc.cache != nil {
		if err := uc.cache.DeleteUserCache(ctx, in.RevieweeID); err != nil {
			log.Printf("reviews: embedding cache delete for %s: %v", in.RevieweeID, err)
		}
	}

	if uc.pointsURL == "" {
		return
	}

	firePointsEvent(uc.pointsURL, in.AuthHeader, map[string]any{
		"userId": in.ReviewerID, "reason": "review_left", "amount": 5, "matchId": in.MatchID,
	})

	tok, err := auth.MintServiceToken(uc.jwtSecret)
	if err != nil {
		log.Printf("reviews: mint service token for reviewee points: %v", err)
		return
	}
	firePointsEvent(uc.pointsURL, "Bearer "+tok, map[string]any{
		"userId": in.RevieweeID, "reason": "review_received", "amount": 3, "matchId": in.MatchID,
	})
}
