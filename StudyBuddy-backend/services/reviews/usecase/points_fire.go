package usecase

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"
)

const defaultSideEffectTimeout = 8 * time.Second

func firePointsEvent(baseURL, authorization string, body map[string]any) {
	if baseURL == "" || authorization == "" {
		return
	}
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), defaultSideEffectTimeout)
		defer cancel()
		payload, err := json.Marshal(body)
		if err != nil {
			return
		}
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimSuffix(baseURL, "/")+"/api/v1/points/events", bytes.NewReader(payload))
		if err != nil {
			return
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", authorization)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			log.Printf("reviews: points event: %v", err)
			return
		}
		resp.Body.Close()
	}()
}
