package usecase

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"studybuddy/backend/pkg/gemini"
)

var ErrSearchQueryRequired = errors.New("search query is required")

// StudentSearchHit is a student profile ranked by semantic similarity to a query.
type StudentSearchHit struct {
	UserID     string
	FirstName  string
	LastName   string
	Bio        string
	AvatarURL  string
	Similarity float64
}

type SearchStudentsInput struct {
	SearcherID string
	Query      string
	Limit      int
}

type SearchStudents interface {
	SearchStudents(ctx context.Context, in SearchStudentsInput) ([]StudentSearchHit, error)
}

type StudentSearchRepository interface {
	SearchByEmbedding(ctx context.Context, excludeUserID string, vector []float32, limit int) ([]StudentSearchHit, error)
}

type searchStudents struct {
	repo  StudentSearchRepository
	embed gemini.Embedder
}

func NewSearchStudents(repo StudentSearchRepository, embed gemini.Embedder) SearchStudents {
	return &searchStudents{repo: repo, embed: embed}
}

func (uc *searchStudents) SearchStudents(ctx context.Context, in SearchStudentsInput) ([]StudentSearchHit, error) {
	query := strings.TrimSpace(in.Query)
	if query == "" {
		return nil, ErrSearchQueryRequired
	}

	limit := in.Limit
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	vec, err := uc.embed.EmbedText(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("embed search query: %w", err)
	}
	if vec == nil {
		return []StudentSearchHit{}, nil
	}

	hits, err := uc.repo.SearchByEmbedding(ctx, in.SearcherID, vec, limit)
	if err != nil {
		return nil, fmt.Errorf("search by embedding: %w", err)
	}
	return hits, nil
}
