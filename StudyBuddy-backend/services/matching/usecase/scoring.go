package usecase

import "math"

const (
	weightAvailability  = 0.30
	weightCourses       = 0.25
	weightInterests     = 0.25
	weightReputation    = 0.10
	weightMutualFriends = 0.10
)

// ComputeOverallScore returns a weighted composite match score in [0, 1].
// Signals: availability overlap (30%), shared courses via Jaccard (25%),
// shared interests via Jaccard (25%), peer reputation (10%), mutual friends (10%).
// A 15% bonus may be applied by the caller when candidates share ≥2 courses
// and more than 30 minutes of availability overlap.
func ComputeOverallScore(availability, courses, interests, reputation, mutualFriends float64) float64 {
	return weightAvailability*availability +
		weightCourses*courses +
		weightInterests*interests +
		weightReputation*reputation +
		weightMutualFriends*mutualFriends
}

// ApplyMatchBonus boosts the overall score when strong course overlap coincides with schedule overlap.
func ApplyMatchBonus(overall float64, sharedCourseCount, overlapMinutes int) float64 {
	if sharedCourseCount >= 2 && overlapMinutes > 30 {
		return math.Min(overall*1.15, 1.0)
	}
	return overall
}

func jaccardScore(a, b []string) float64 {
	if len(a) == 0 && len(b) == 0 {
		return 0
	}
	setA := toSet(a)
	setB := toSet(b)
	intersection := 0
	for k := range setA {
		if setB[k] {
			intersection++
		}
	}
	union := len(setA) + len(setB) - intersection
	if union == 0 {
		return 0
	}
	return float64(intersection) / float64(union)
}
