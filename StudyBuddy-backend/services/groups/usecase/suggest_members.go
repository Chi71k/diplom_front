package usecase

import (
	"context"
	"math"
	"sort"
	"strings"

	"studybuddy/backend/pkg/embedding"
	"studybuddy/backend/services/groups/domain"
)

type SuggestMembersForGroupInput struct {
	GroupID string
	Limit   int
}

type MemberSuggestion struct {
	UserID          string
	FirstName       string
	LastName        string
	AvatarURL       string
	SimilarityScore float64
}

type SuggestMembersForGroup interface {
	SuggestMembersForGroup(ctx context.Context, in SuggestMembersForGroupInput) ([]MemberSuggestion, error)
}

type suggestMembersForGroup struct {
	repo   GroupRepository
	embed  EmbeddingProvider
	gemini *embedding.Client
}

func NewSuggestMembersForGroup(repo GroupRepository, ep EmbeddingProvider, gemini *embedding.Client) SuggestMembersForGroup {
	return &suggestMembersForGroup{repo: repo, embed: ep, gemini: gemini}
}

func buildGroupProfileText(name, description string, courseTitles, interestNames []string) string {
	var b strings.Builder
	b.WriteString("Group: ")
	b.WriteString(strings.TrimSpace(name))
	b.WriteString(". Description: ")
	b.WriteString(strings.TrimSpace(description))
	b.WriteString(". Courses: ")
	b.WriteString(strings.Join(courseTitles, ", "))
	b.WriteString(". Current members interests: ")
	b.WriteString(strings.Join(interestNames, ", "))
	b.WriteString(".")
	return b.String()
}

func (uc *suggestMembersForGroup) SuggestMembersForGroup(ctx context.Context, in SuggestMembersForGroupInput) ([]MemberSuggestion, error) {
	limit := in.Limit
	if limit <= 0 {
		limit = 10
	}

	g, err := uc.repo.GetByID(ctx, in.GroupID)
	if err != nil {
		return nil, err
	}
	if g == nil {
		return nil, domain.ErrGroupNotFound
	}

	titles, err := uc.repo.ListCourseTitlesForGroup(ctx, in.GroupID)
	if err != nil {
		return nil, err
	}
	interests, err := uc.repo.ListDistinctInterestNamesForGroup(ctx, in.GroupID)
	if err != nil {
		return nil, err
	}

	text := buildGroupProfileText(g.Name, g.Description, titles, interests)
	groupVec, err := uc.gemini.Embed(ctx, text)
	if err != nil || groupVec == nil || len(groupVec) == 0 {
		return uc.fallbackOverlap(ctx, in.GroupID, limit, titles)
	}

	candidates, err := uc.repo.ListCandidateUserIDs(ctx, in.GroupID, 200)
	if err != nil {
		return nil, err
	}

	type scored struct {
		id    string
		score float64
	}
	var rows []scored
	for _, uid := range candidates {
		cv, err := uc.embed.GetOrCompute(ctx, uid)
		if err != nil {
			continue
		}
		if cv == nil || len(cv) != len(groupVec) {
			continue
		}
		s := embedding.CosineSimilarity(groupVec, cv)
		rows = append(rows, scored{id: uid, score: s})
	}

	if len(rows) == 0 {
		return uc.fallbackOverlap(ctx, in.GroupID, limit, titles)
	}

	sort.Slice(rows, func(i, j int) bool { return rows[i].score > rows[j].score })
	if len(rows) > limit {
		rows = rows[:limit]
	}

	ids := make([]string, len(rows))
	for i := range rows {
		ids[i] = rows[i].id
	}
	profiles, err := uc.repo.ListProfiles(ctx, ids)
	if err != nil {
		return nil, err
	}
	byID := make(map[string]ProfileSnippet, len(profiles))
	for _, p := range profiles {
		byID[p.UserID] = p
	}

	out := make([]MemberSuggestion, 0, len(rows))
	for _, r := range rows {
		p := byID[r.id]
		out = append(out, MemberSuggestion{
			UserID:          r.id,
			FirstName:       p.FirstName,
			LastName:        p.LastName,
			AvatarURL:       p.AvatarURL,
			SimilarityScore: r.score,
		})
	}
	return out, nil
}

func (uc *suggestMembersForGroup) fallbackOverlap(ctx context.Context, groupID string, limit int, groupCourseTitles []string) ([]MemberSuggestion, error) {
	overlaps, err := uc.repo.ListCourseOverlapCandidates(ctx, groupID, max(200, limit))
	if err != nil {
		return nil, err
	}
	if len(overlaps) == 0 {
		return []MemberSuggestion{}, nil
	}

	denom := float64(len(groupCourseTitles))
	if denom < 1 {
		denom = 1
	}

	sort.Slice(overlaps, func(i, j int) bool {
		if overlaps[i].OverlapCount == overlaps[j].OverlapCount {
			return overlaps[i].UserID < overlaps[j].UserID
		}
		return overlaps[i].OverlapCount > overlaps[j].OverlapCount
	})
	if len(overlaps) > limit {
		overlaps = overlaps[:limit]
	}

	ids := make([]string, len(overlaps))
	for i := range overlaps {
		ids[i] = overlaps[i].UserID
	}
	profiles, err := uc.repo.ListProfiles(ctx, ids)
	if err != nil {
		return nil, err
	}
	byID := make(map[string]ProfileSnippet, len(profiles))
	for _, p := range profiles {
		byID[p.UserID] = p
	}

	out := make([]MemberSuggestion, 0, len(overlaps))
	for _, o := range overlaps {
		p := byID[o.UserID]
		score := math.Min(float64(o.OverlapCount)/denom, 1.0)
		out = append(out, MemberSuggestion{
			UserID:          o.UserID,
			FirstName:       p.FirstName,
			LastName:        p.LastName,
			AvatarURL:       p.AvatarURL,
			SimilarityScore: score,
		})
	}
	return out, nil
}
