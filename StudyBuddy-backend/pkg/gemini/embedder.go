// Package gemini provides a client for generating text embeddings
// using the Google Gemini API (model: text-embedding-004).
package gemini

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const (
	embedModel       = "models/gemini-embedding-2"
	embedURL         = "https://generativelanguage.googleapis.com/v1beta/models/gemini-embedding-2:embedContent"
	embeddingDims    = 768
	embedHTTPTimeout = 10 * time.Second
)

// Embedder generates 768-dimensional embedding vectors for text.
type Embedder interface {
	// EmbedText returns a 768-dimensional vector for the given text.
	// Returns an error if the API call fails or the response is malformed.
	EmbedText(ctx context.Context, text string) ([]float32, error)
}

type httpEmbedder struct {
	apiKey     string
	httpClient *http.Client
}

// NewEmbedder constructs an Embedder backed by the Gemini API.
// apiKey must be a valid Google AI Studio or Vertex AI API key.
func NewEmbedder(apiKey string) (Embedder, error) {
	if strings.TrimSpace(apiKey) == "" {
		return nil, fmt.Errorf("gemini api key is required")
	}
	return &httpEmbedder{
		apiKey:     apiKey,
		httpClient: &http.Client{Timeout: embedHTTPTimeout},
	}, nil
}

// NoOpEmbedder returns nil vectors without error (semantic scoring uses neutral fallback).
type NoOpEmbedder struct{}

func (NoOpEmbedder) EmbedText(context.Context, string) ([]float32, error) {
	return nil, nil
}

type embedRequest struct {
	Model               string         `json:"model"`
	Content             contentWrapper `json:"content"`
	TaskType            string         `json:"taskType,omitempty"`
	OutputDimensionality int           `json:"outputDimensionality,omitempty"`
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

func (c *httpEmbedder) EmbedText(ctx context.Context, text string) ([]float32, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil, fmt.Errorf("embed text is empty")
	}

	ctx, cancel := context.WithTimeout(ctx, embedHTTPTimeout)
	defer cancel()

	body := embedRequest{
		Model:                embedModel,
		TaskType:             "SEMANTIC_SIMILARITY",
		OutputDimensionality: embeddingDims,
	}
	body.Content.Parts = []struct {
		Text string `json:"text"`
	}{{Text: text}}

	payload, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal embed request: %w", err)
	}

	url := embedURL + "?key=" + c.apiKey
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
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
	if len(out.Embedding.Values) != embeddingDims {
		return nil, fmt.Errorf("expected %d dimensions, got %d", embeddingDims, len(out.Embedding.Values))
	}

	vec := make([]float32, embeddingDims)
	for i, v := range out.Embedding.Values {
		vec[i] = float32(v)
	}
	return vec, nil
}
