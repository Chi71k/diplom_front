package usecase

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"studybuddy/backend/pkg/auth"
)

// FireGroupCreatedPoints posts a points event for the group owner (fire-and-forget).
func FireGroupCreatedPoints(pointsBaseURL string, jwtSecret []byte, ownerID, groupID string) {
	base := strings.TrimSuffix(strings.TrimSpace(pointsBaseURL), "/")
	if base == "" || ownerID == "" {
		return
	}
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		tok, err := auth.MintServiceToken(jwtSecret)
		if err != nil {
			log.Printf("points mint group_created: %v", err)
			return
		}
		body, err := json.Marshal(map[string]any{
			"userId":  ownerID,
			"reason":  "group_created",
			"amount":  10,
			"matchId": groupID,
		})
		if err != nil {
			return
		}
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, base+"/api/v1/points/events", bytes.NewReader(body))
		if err != nil {
			return
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+tok)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			log.Printf("points event group_created: %v", err)
			return
		}
		resp.Body.Close()
	}()
}
