package usecase

import (
	"context"
	"fmt"
	"log"
	"math"
	"sort"
	"studybuddy/backend/pkg/embedding"
	"studybuddy/backend/services/matching/domain"
	"time"
)

const (
	weightSemantic      = 0.35
	weightAvailability  = 0.30
	weightCourses       = 0.20
	weightReputation    = 0.10
	weightMutualFriends = 0.05
)

type ListCandidatesInput struct {
	RequesterID string
	Limit       int // defaults to 20
}

type ListCandidates interface {
	ListCandidates(ctx context.Context, in ListCandidatesInput) ([]domain.MatchCandidate, error)
}

type listCandidates struct {
	matches     MatchRepository
	profiles    ProfileClient
	slots       SlotClient
	courses     CourseClient
	candidates  CandidateStore
	embed       EmbeddingProvider
	reputation  ReputationClient
	friendships FriendshipRepository
}

func NewListCandidates(
	matches MatchRepository,
	profiles ProfileClient,
	slots SlotClient,
	courses CourseClient,
	candidates CandidateStore,
	embed EmbeddingProvider,
	reputation ReputationClient,
	friendships FriendshipRepository,
) ListCandidates {
	return &listCandidates{
		matches:     matches,
		profiles:    profiles,
		slots:       slots,
		courses:     courses,
		candidates:  candidates,
		embed:       embed,
		reputation:  reputation,
		friendships: friendships,
	}
}

func (uc *listCandidates) ListCandidates(ctx context.Context, in ListCandidatesInput) ([]domain.MatchCandidate, error) {
	limit := in.Limit
	if limit <= 0 {
		limit = 20
	}

	// Excluding already matched students from candidates
	existing, err := uc.matches.ListForUser(ctx, in.RequesterID, ListMatchesFilter{Limit: 1000})
	if err != nil {
		return nil, fmt.Errorf("list existing matches: %w", err)
	}
	excludeIDs := make([]string, 0, len(existing))
	for _, m := range existing {
		other := m.ReceiverID
		if m.RequesterID != in.RequesterID {
			other = m.RequesterID
		}
		excludeIDs = append(excludeIDs, other)
	}

	candidateIDs, err := uc.candidates.ListCandidateIDs(ctx, in.RequesterID, excludeIDs)
	if err != nil {
		return nil, fmt.Errorf("list candidate ids: %w", err)
	}
	if len(candidateIDs) == 0 {
		return []domain.MatchCandidate{}, nil
	}

	// Requester's data for scoring.
	myInterests, err := uc.profiles.GetInterestIDs(ctx, in.RequesterID)
	if err != nil {
		return nil, fmt.Errorf("get my interests: %w", err)
	}
	myCourses, err := uc.courses.ListCourseIDsForUser(ctx, in.RequesterID)
	if err != nil {
		return nil, fmt.Errorf("get my courses: %w", err)
	}
	mySlots, err := uc.slots.ListForUser(ctx, in.RequesterID)
	if err != nil {
		return nil, fmt.Errorf("get my slots: %w", err)
	}

	reqEmb, _ := uc.embed.GetOrCompute(ctx, in.RequesterID)

	// Batch-fetch candidate slots.
	allSlots, err := uc.slots.ListForUsers(ctx, candidateIDs)
	if err != nil {
		return nil, fmt.Errorf("batch-fetch candidate slots: %w", err)
	}
	slotsByUser := groupSlotsByUser(allSlots)

	var fallbackUsed bool

	result := make([]domain.MatchCandidate, 0, len(candidateIDs))
	for _, cid := range candidateIDs {
		profile, err := uc.profiles.GetProfile(ctx, cid)
		if err != nil || profile == nil {
			continue
		}

		candEmb, _ := uc.embed.GetOrCompute(ctx, cid)

		var interestJaccard float64
		if candEmb == nil {
			theirInterests, ierr := uc.profiles.GetInterestIDs(ctx, cid)
			if ierr != nil {
				theirInterests = nil
			}
			interestJaccard = jaccardScore(myInterests, theirInterests)
		}

		theirCourses, _ := uc.courses.ListCourseIDsForUser(ctx, cid)
		theirSlots := slotsByUser[cid]

		commonCourses := intersectStrings(myCourses, theirCourses)
		overlaps := computeOverlaps(mySlots, theirSlots)
		availScore := availabilityScore(mySlots, theirSlots)
		courseJaccard := jaccardScore(myCourses, theirCourses)

		var semanticForScore float64
		var semanticDisplay float64
		if reqEmb != nil && candEmb != nil {
			semanticForScore = embedding.CosineSimilarity(reqEmb, candEmb)
			semanticDisplay = semanticForScore
		} else {
			fallbackUsed = true
			semanticForScore = interestJaccard
			semanticDisplay = interestJaccard
		}

		repScore := uc.reputation.GetAverageRating(ctx, cid)

		mutuals, err := uc.friendships.MutualFriendCount(ctx, in.RequesterID, cid)
		if err != nil {
			mutuals = 0
		}
		mutualNorm := math.Min(float64(mutuals)/10.0, 1.0)

		score := weightSemantic*semanticForScore +
			weightAvailability*availScore +
			weightCourses*courseJaccard +
			weightReputation*repScore +
			weightMutualFriends*mutualNorm

		overlapMin := totalOverlapMinutes(mySlots, theirSlots)
		if len(commonCourses) >= 2 && overlapMin > 30 {
			score = math.Min(score*1.15, 1.0)
		}

		interestScoreOut := interestJaccard
		if candEmb != nil {
			// Embedding path: cosine (or same value used for semantic weight) is the interest-compatibility signal.
			interestScoreOut = semanticDisplay
		}

		result = append(result, domain.MatchCandidate{
			UserID:        profile.UserID,
			FirstName:     profile.FirstName,
			LastName:      profile.LastName,
			Bio:           profile.Bio,
			AvatarURL:     profile.AvatarURL,
			CommonCourses: commonCourses,
			CommonSlots:   overlaps,
			SemanticScore: semanticDisplay,
			InterestScore: interestScoreOut,
			AvailScore:    availScore,
			OverallScore:  score,
		})
	}

	if fallbackUsed {
		log.Printf("warning: list candidates used interest overlap fallback for semantic scoring (embeddings unavailable)")
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].OverallScore > result[j].OverallScore
	})

	applyDiversityFloor(result)

	if limit < len(result) {
		result = result[:limit]
	}
	return result, nil
}

func applyDiversityFloor(cands []domain.MatchCandidate) {
	if len(cands) < 5 {
		return
	}
	for i := 0; i < 5; i++ {
		if cands[i].SemanticScore <= 0.92 {
			return
		}
	}
	var replacement *domain.MatchCandidate
	best := -1.0
	for i := 5; i < len(cands); i++ {
		if cands[i].SemanticScore <= 0.92 && cands[i].OverallScore > best {
			best = cands[i].OverallScore
			cp := cands[i]
			replacement = &cp
		}
	}
	if replacement != nil {
		cands[4] = *replacement
	}
}

// helpers
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

func availabilityScore(mine, theirs []SlotData) float64 {
	if len(mine) == 0 || len(theirs) == 0 {
		return 0
	}
	myMinutes := totalMinutes(mine)
	if myMinutes == 0 {
		return 0
	}
	overlapMinutes := totalOverlapMinutes(mine, theirs)
	score := float64(overlapMinutes) / float64(myMinutes)
	if score > 1 {
		score = 1
	}
	return score
}

// computeOverlaps returns the list of overlapping time windows.
func computeOverlaps(mine, theirs []SlotData) []domain.SlotOverlap {
	var out []domain.SlotOverlap
	for _, m := range mine {
		for _, t := range theirs {
			if m.DayOfWeek != t.DayOfWeek {
				continue
			}
			mStart, mEnd := parseHHMM(m.StartTime), parseHHMM(m.EndTime)
			tStart, tEnd := parseHHMM(t.StartTime), parseHHMM(t.EndTime)
			oStart := maxInt(mStart, tStart)
			oEnd := minInt(mEnd, tEnd)
			if oEnd > oStart {
				out = append(out, domain.SlotOverlap{
					DayOfWeek: m.DayOfWeek,
					StartTime: formatHHMM(oStart),
					EndTime:   formatHHMM(oEnd),
					Timezone:  m.Timezone,
				})
			}
		}
	}
	return out
}

func totalMinutes(slots []SlotData) int {
	total := 0
	for _, s := range slots {
		total += parseHHMM(s.EndTime) - parseHHMM(s.StartTime)
	}
	return total
}

func totalOverlapMinutes(mine, theirs []SlotData) int {
	total := 0
	for _, m := range mine {
		for _, t := range theirs {
			if m.DayOfWeek != t.DayOfWeek {
				continue
			}
			oStart := maxInt(parseHHMM(m.StartTime), parseHHMM(t.StartTime))
			oEnd := minInt(parseHHMM(m.EndTime), parseHHMM(t.EndTime))
			if oEnd > oStart {
				total += oEnd - oStart
			}
		}
	}
	return total
}

// parseHHMM converts "HH:MM" → total minutes since midnight.
func parseHHMM(s string) int {
	t, err := time.Parse("15:04", s)
	if err != nil {
		return 0
	}
	return t.Hour()*60 + t.Minute()
}

func formatHHMM(minutes int) string {
	h := minutes / 60
	m := minutes % 60
	return fmt.Sprintf("%02d:%02d", h, m)
}

func groupSlotsByUser(slots []SlotData) map[string][]SlotData {
	out := make(map[string][]SlotData)
	for _, s := range slots {
		out[s.UserID] = append(out[s.UserID], s)
	}
	return out
}

func toSet(ids []string) map[string]bool {
	s := make(map[string]bool, len(ids))
	for _, id := range ids {
		s[id] = true
	}
	return s
}

func intersectStrings(a, b []string) []string {
	sb := toSet(b)
	var out []string
	for _, v := range a {
		if sb[v] {
			out = append(out, v)
		}
	}
	return out
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
