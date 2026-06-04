package usecase

import "math"

// CosineSimilarity computes the cosine similarity between two
// equal-length float32 vectors. Returns 0.5 (neutral) if either
// vector is nil or zero-magnitude.
func CosineSimilarity(a, b []float32) float64 {
	if len(a) == 0 || len(b) == 0 || len(a) != len(b) {
		return 0.5
	}
	var dot, na, nb float64
	for i := range a {
		dot += float64(a[i]) * float64(b[i])
		na += float64(a[i]) * float64(a[i])
		nb += float64(b[i]) * float64(b[i])
	}
	if na == 0 || nb == 0 {
		return 0.5
	}
	return dot / (math.Sqrt(na) * math.Sqrt(nb))
}

// ScoreSemantic returns the semantic component score [0,1].
// Returns 0.5 if either vector is nil (graceful degradation).
func ScoreSemantic(requesterVec, candidateVec []float32) float64 {
	if len(requesterVec) == 0 || len(candidateVec) == 0 {
		return 0.5
	}
	return CosineSimilarity(requesterVec, candidateVec)
}
