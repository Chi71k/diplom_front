package embedding

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strings"
	"time"
)

const geminiEmbedURL = "https://generativelanguage.googleapis.com/v1beta/models/text-embedding-004:embedContent"

type Client struct {
	apiKey     string
	httpClient *http.Client
}

func NewClient(apiKey string) *Client {
	return &Client{apiKey: apiKey, httpClient: &http.Client{}}
}

type embedRequest struct {
	Model    string         `json:"model"`
	Content  contentWrapper `json:"content"`
	TaskType string         `json:"taskType,omitempty"`
}

type contentWrapper struct {
	Parts []struct {
		Text string `json:"text"`
	} `json:"parts"`
}

type embedResponse struct {
	Embedding *struct {
		Values []float64 `json:"values"`
	} `json:"embedding"`
}

// Embed returns a 768-dimensional embedding vector for the given text.
// Returns nil, nil if text is empty (caller must handle gracefully).
func (c *Client) Embed(ctx context.Context, text string) ([]float64, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil, nil
	}

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	body := embedRequest{
		Model:    "models/text-embedding-004",
		TaskType: "SEMANTIC_SIMILARITY",
	}
	body.Content.Parts = []struct {
		Text string `json:"text"`
	}{{Text: text}}

	payload, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal embed request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.embedURL(), bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("gemini embed: status %d", resp.StatusCode)
	}

	var out embedResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("decode embed response: %w", err)
	}
	if out.Embedding == nil || len(out.Embedding.Values) == 0 {
		return nil, fmt.Errorf("embed response missing embedding.values")
	}
	return out.Embedding.Values, nil
}

func (c *Client) embedURL() string {
	if c.apiKey == "" {
		return geminiEmbedURL
	}
	return geminiEmbedURL + "?key=" + c.apiKey
}

// CosineSimilarity returns the cosine similarity between two equal-length vectors.
// Returns 0 if either vector is nil or zero-length.
func CosineSimilarity(a, b []float64) float64 {
	if len(a) == 0 || len(b) == 0 || len(a) != len(b) {
		return 0
	}
	var dot, na, nb float64
	for i := range a {
		dot += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	if na == 0 || nb == 0 {
		return 0
	}
	return dot / (math.Sqrt(na) * math.Sqrt(nb))
}

// BuildUserProfileText builds the text that represents a user for embedding.
func BuildUserProfileText(firstName, lastName, bio string, interests, courseTitles []string) string {
	var b strings.Builder
	name := strings.TrimSpace(strings.Join([]string{strings.TrimSpace(firstName), strings.TrimSpace(lastName)}, " "))
	if name != "" {
		b.WriteString(name)
		b.WriteString(". ")
	}
	if s := strings.TrimSpace(bio); s != "" {
		b.WriteString("Bio: ")
		b.WriteString(s)
		b.WriteString(". ")
	}
	if len(interests) > 0 {
		clean := make([]string, 0, len(interests))
		for _, i := range interests {
			if t := strings.TrimSpace(i); t != "" {
				clean = append(clean, t)
			}
		}
		if len(clean) > 0 {
			b.WriteString("Interests: ")
			b.WriteString(strings.Join(clean, ", "))
			b.WriteString(". ")
		}
	}
	if len(courseTitles) > 0 {
		clean := make([]string, 0, len(courseTitles))
		for _, t := range courseTitles {
			if s := strings.TrimSpace(t); s != "" {
				clean = append(clean, s)
			}
		}
		if len(clean) > 0 {
			b.WriteString("Courses: ")
			b.WriteString(strings.Join(clean, ", "))
			b.WriteString(". ")
		}
	}
	return strings.TrimSpace(b.String())
}
