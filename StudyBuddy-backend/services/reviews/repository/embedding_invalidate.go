package repository

import (
	"context"

	"studybuddy/backend/pkg/embedding"
)

// EmbeddingCacheInvalidatorAdapter wraps PgCache for reviews use case.
type EmbeddingCacheInvalidatorAdapter struct {
	cache *embedding.PgCache
}

func NewEmbeddingCacheInvalidatorAdapter(cache *embedding.PgCache) *EmbeddingCacheInvalidatorAdapter {
	return &EmbeddingCacheInvalidatorAdapter{cache: cache}
}

func (a *EmbeddingCacheInvalidatorAdapter) DeleteUserCache(ctx context.Context, userID string) error {
	if a == nil || a.cache == nil {
		return nil
	}
	return a.cache.Delete(ctx, userID)
}
