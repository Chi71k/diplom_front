package usecase

import (
	"fmt"
	"strings"
)

// BuildNarrative composes a natural-language profile description
// from a user's structured data for submission to the Gemini
// Embedding API.
//
// Format:
// "Student [FirstName] [LastName] is interested in [interests joined
// by comma]. They are enrolled in [courses joined by comma].
// About them: [bio]."
//
// Rules:
// - If bio is empty, omit the "About them:" sentence.
// - If interests slice is empty, omit that sentence.
// - If courses slice is empty, omit that sentence.
// - Never return an empty string — always include the name line.
func BuildNarrative(firstName, lastName string, interests, courses []string, bio string) string {
	name := strings.TrimSpace(strings.Join([]string{strings.TrimSpace(firstName), strings.TrimSpace(lastName)}, " "))
	if name == "" {
		name = "Student"
	} else {
		name = "Student " + name
	}

	var parts []string
	parts = append(parts, name)

	if cleaned := joinNonEmpty(interests); cleaned != "" {
		parts = append(parts, fmt.Sprintf("is interested in %s.", cleaned))
	}
	if cleaned := joinNonEmpty(courses); cleaned != "" {
		parts = append(parts, fmt.Sprintf("They are enrolled in %s.", cleaned))
	}
	if s := strings.TrimSpace(bio); s != "" {
		parts = append(parts, fmt.Sprintf("About them: %s.", s))
	}

	return strings.Join(parts, " ")
}

func joinNonEmpty(items []string) string {
	clean := make([]string, 0, len(items))
	for _, item := range items {
		if s := strings.TrimSpace(item); s != "" {
			clean = append(clean, s)
		}
	}
	return strings.Join(clean, ", ")
}
