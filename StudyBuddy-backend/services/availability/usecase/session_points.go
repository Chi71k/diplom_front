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

// FireSessionConfirmedPoints posts points for each participant (fire-and-forget).
func FireSessionConfirmedPoints(pointsServiceURL string, jwtSecret []byte, sessionID string, participantUserIDs []string) {
	base := strings.TrimSuffix(strings.TrimSpace(pointsServiceURL), "/")
	if base == "" {
		return
	}
	for _, uid := range participantUserIDs {
		if uid == "" {
			continue
		}
		u := uid
		sid := sessionID
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			tok, err := auth.MintServiceToken(jwtSecret)
			if err != nil {
				log.Printf("points mint session_confirmed: %v", err)
				return
			}
			body, err := json.Marshal(map[string]any{
				"userId":  u,
				"reason":  "session_confirmed",
				"amount":  5,
				"matchId": sid,
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
				log.Printf("points event session_confirmed: %v", err)
				return
			}
			resp.Body.Close()
		}()
	}
}
