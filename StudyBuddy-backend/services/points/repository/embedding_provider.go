package repository

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"log"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"

	"studybuddy/backend/pkg/embedding"
	"studybuddy/backend/services/points/usecase"
)

type geminiEmbeddingProvider struct {
	client *embedding.Client
	cache  embedding.Cache
	src    *userEmbedSource
}

func NewEmbeddingProvider(client *embedding.Client, cache embedding.Cache, pool *pgxpool.Pool) usecase.EmbeddingProvider {
	return &geminiEmbeddingProvider{
		client: client,
		cache:  cache,
		src:    newUserEmbedSource(pool),
	}
}

func (p *geminiEmbeddingProvider) GetOrCompute(ctx context.Context, userID string) ([]float64, error) {
	prof, err := p.src.GetProfile(ctx, userID)
	if err != nil || prof == nil {
		return nil, nil
	}
	names, err := p.src.ListInterestNames(ctx, userID)
	if err != nil {
		log.Printf("points embedding: interest names for %s: %v", userID, err)
		names = nil
	}
	titles, err := p.src.ListCourseTitles(ctx, userID)
	if err != nil {
		log.Printf("points embedding: course titles for %s: %v", userID, err)
		titles = nil
	}
	text := embedding.BuildUserProfileText(prof.FirstName, prof.LastName, prof.Bio, names, titles)
	text = strings.TrimSpace(text)
	if text == "" {
		return nil, nil
	}
	sum := sha256.Sum256([]byte(text))
	hash := hex.EncodeToString(sum[:])
	cached, cachedHash, err := p.cache.Get(ctx, userID)
	if err != nil {
		log.Printf("points embedding cache get: %v", err)
	} else if len(cached) > 0 && cachedHash == hash {
		return cached, nil
	}
	vec, err := p.client.Embed(ctx, text)
	if err != nil {
		log.Printf("points gemini embed failed for user %s: %v", userID, err)
		return nil, nil
	}
	if vec == nil {
		return nil, nil
	}
	if err := p.cache.Set(ctx, userID, vec, hash); err != nil {
		log.Printf("points embedding cache set for user %s: %v", userID, err)
	}
	return vec, nil
}
