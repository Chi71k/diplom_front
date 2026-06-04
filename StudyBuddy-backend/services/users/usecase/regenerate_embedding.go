package usecase

import (
	"context"
	"log"
	"time"

	"studybuddy/backend/pkg/embedding"
	"studybuddy/backend/pkg/gemini"
)

// EmbeddingRegenerator triggers async embedding regeneration for a user.
type EmbeddingRegenerator interface {
	RegenerateAsync(userID string)
}

type embeddingRegenerator struct {
	profiles ProfileEmbeddingSource
	embed    gemini.Embedder
	repo     EmbeddingRepository
	cache    embedding.Cache
}

// NewEmbeddingRegenerator creates a regenerator that updates stored vectors in the background.
func NewEmbeddingRegenerator(
	profiles ProfileEmbeddingSource,
	embed gemini.Embedder,
	repo EmbeddingRepository,
	cache embedding.Cache,
) EmbeddingRegenerator {
	return &embeddingRegenerator{
		profiles: profiles,
		embed:    embed,
		repo:     repo,
		cache:    cache,
	}
}

// RegenerateAsync fires embedding generation in a background goroutine.
func (r *embeddingRegenerator) RegenerateAsync(userID string) {
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		prof, err := r.profiles.GetByUserID(ctx, userID)
		if err != nil || prof == nil {
			log.Printf("embedding regenerate: profile for %s: %v", userID, err)
			return
		}

		interests, err := r.profiles.ListInterestNamesForUser(ctx, userID)
		if err != nil {
			log.Printf("embedding regenerate: interests for %s: %v", userID, err)
			interests = nil
		}
		courses, err := r.profiles.ListCourseTitlesForUser(ctx, userID)
		if err != nil {
			log.Printf("embedding regenerate: courses for %s: %v", userID, err)
			courses = nil
		}

		narrative := BuildNarrative(prof.FirstName, prof.LastName, interests, courses, prof.Bio)
		vec, err := r.embed.EmbedText(ctx, narrative)
		if err != nil {
			log.Printf("embedding regenerate: gemini for %s: %v", userID, err)
			return
		}
		if vec == nil {
			return
		}

		if err := r.repo.UpdateEmbedding(ctx, userID, vec); err != nil {
			log.Printf("embedding regenerate: persist for %s: %v", userID, err)
			return
		}

		if r.cache != nil {
			if err := r.cache.Delete(ctx, userID); err != nil {
				log.Printf("embedding regenerate: cache invalidate for %s: %v", userID, err)
			}
		}
	}()
}
