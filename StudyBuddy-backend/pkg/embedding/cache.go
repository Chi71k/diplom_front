package embedding

import "context"

// Cache persists embedding vectors keyed by user and profile hash.
type Cache interface {
	Get(ctx context.Context, userID string) ([]float64, string, error)
	Set(ctx context.Context, userID string, embedding []float64, profileHash string) error
	Delete(ctx context.Context, userID string) error
}
