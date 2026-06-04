package usecase

// EmbeddingHookAdapter bridges auth registration to users embedding regeneration.
type EmbeddingHookAdapter struct {
	Regenerate func(userID string)
}

func (h *EmbeddingHookAdapter) OnUserRegistered(userID string) {
	if h != nil && h.Regenerate != nil {
		h.Regenerate(userID)
	}
}
