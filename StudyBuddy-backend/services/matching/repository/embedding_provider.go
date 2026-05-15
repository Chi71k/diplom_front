package repository

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"log"
	"strings"

	"studybuddy/backend/pkg/embedding"
	"studybuddy/backend/services/matching/usecase"
)

type geminiEmbeddingProvider struct {
	client   *embedding.Client
	cache    embedding.Cache
	profiles usecase.ProfileClient
	courses  usecase.CourseClient
}

func NewEmbeddingProvider(
	client *embedding.Client,
	cache embedding.Cache,
	profiles usecase.ProfileClient,
	courses usecase.CourseClient,
) usecase.EmbeddingProvider {
	return &geminiEmbeddingProvider{
		client:   client,
		cache:    cache,
		profiles: profiles,
		courses:  courses,
	}
}

func (p *geminiEmbeddingProvider) GetOrCompute(ctx context.Context, userID string) ([]float64, error) {
	prof, err := p.profiles.GetProfile(ctx, userID)
	if err != nil || prof == nil {
		return nil, nil
	}
	names, err := p.profiles.ListInterestNamesForUser(ctx, userID)
	if err != nil {
		log.Printf("embedding: interest names for %s: %v", userID, err)
		names = nil
	}
	titles, err := p.courses.ListCourseTitlesForUser(ctx, userID)
	if err != nil {
		log.Printf("embedding: course titles for %s: %v", userID, err)
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
		log.Printf("embedding cache get: %v", err)
	} else if len(cached) > 0 && cachedHash == hash {
		return cached, nil
	}

	vec, err := p.client.Embed(ctx, text)
	if err != nil {
		log.Printf("gemini embed failed for user %s: %v", userID, err)
		return nil, nil
	}
	if vec == nil {
		return nil, nil
	}

	if err := p.cache.Set(ctx, userID, vec, hash); err != nil {
		log.Printf("embedding cache set for user %s: %v", userID, err)
	}
	return vec, nil
}
