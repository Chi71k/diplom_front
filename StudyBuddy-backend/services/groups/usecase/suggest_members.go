package usecase

import (
	"context"
	"math"
	"sort"

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
	repo GroupRepository
}

func NewSuggestMembersForGroup(repo GroupRepository) SuggestMembersForGroup {
	return &suggestMembersForGroup{repo: repo}
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

	return uc.suggestByCourseOverlap(ctx, in.GroupID, limit, titles)
}

func (uc *suggestMembersForGroup) suggestByCourseOverlap(ctx context.Context, groupID string, limit int, groupCourseTitles []string) ([]MemberSuggestion, error) {
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
