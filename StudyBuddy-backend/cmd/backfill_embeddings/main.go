// Command backfill_embeddings generates and stores Gemini embeddings for
// existing users who have NULL embedding_vector (one-off migration helper).
package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/joho/godotenv"

	"studybuddy/backend/pkg/db"
	"studybuddy/backend/pkg/gemini"
	"studybuddy/backend/services/users/repository"
	"studybuddy/backend/services/users/usecase"
)

const (
	batchSize      = 10
	batchPause     = 100 * time.Millisecond
	userTimeout    = 60 * time.Second
	connectTimeout = 10 * time.Second
)

func main() {
	_ = godotenv.Load(".env")

	geminiKey := strings.TrimSpace(os.Getenv("GEMINI_API_KEY"))
	if geminiKey == "" {
		fmt.Fprintln(os.Stderr, "GEMINI_API_KEY is required for backfill — set it and retry")
		os.Exit(1)
	}

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://studybuddy:studybuddy@localhost:5432/studybuddy?sslmode=disable"
	}

	ctx, cancel := context.WithTimeout(context.Background(), connectTimeout)
	pool, err := db.NewPool(ctx, dsn)
	cancel()
	if err != nil {
		fmt.Fprintf(os.Stderr, "database connection failed: %v\n", err)
		os.Exit(1)
	}
	defer pool.Close()

	embedder, err := gemini.NewEmbedder(geminiKey)
	if err != nil {
		fmt.Fprintf(os.Stderr, "gemini embedder: %v\n", err)
		os.Exit(1)
	}

	repo := repository.NewPgProfileRepository(pool)

	listCtx, listCancel := context.WithTimeout(context.Background(), 30*time.Second)
	users, err := repo.ListUsersNeedingEmbeddingBackfill(listCtx)
	listCancel()
	if err != nil {
		fmt.Fprintf(os.Stderr, "list users for backfill: %v\n", err)
		os.Exit(1)
	}

	total := len(users)
	if total == 0 {
		fmt.Println("No users need backfill (all active users already have embedding_vector).")
		os.Exit(0)
	}

	fmt.Printf("Backfilling embeddings for %d user(s)...\n", total)

	var succeeded, failed int
	var mu sync.Mutex

	for batchStart := 0; batchStart < total; batchStart += batchSize {
		batchEnd := batchStart + batchSize
		if batchEnd > total {
			batchEnd = total
		}
		batch := users[batchStart:batchEnd]

		var wg sync.WaitGroup
		sem := make(chan struct{}, batchSize)

		for i, u := range batch {
			wg.Add(1)
			sem <- struct{}{}
			idx := batchStart + i + 1
			user := u
			go func() {
				defer wg.Done()
				defer func() { <-sem }()

				err := backfillUser(context.Background(), repo, embedder, user)
				mu.Lock()
				if err != nil {
					failed++
					fmt.Printf("[%d/%d] user %s — ERROR: %v (skipped, continuing)\n", idx, total, user.ID, err)
				} else {
					succeeded++
					fmt.Printf("[%d/%d] user %s — OK\n", idx, total, user.ID)
				}
				mu.Unlock()
			}()
		}
		wg.Wait()

		if batchEnd < total {
			time.Sleep(batchPause)
		}
	}

	fmt.Printf("\nBackfill complete: %d succeeded, %d failed.\n", succeeded, failed)
	if failed > 0 {
		os.Exit(1)
	}
}

func backfillUser(ctx context.Context, repo *repository.PgProfileRepository, embedder gemini.Embedder, user repository.BackfillUser) error {
	ctx, cancel := context.WithTimeout(ctx, userTimeout)
	defer cancel()

	interests, err := repo.ListInterestNamesForUser(ctx, user.ID)
	if err != nil {
		return fmt.Errorf("interests: %w", err)
	}
	courses, err := repo.ListCourseTitlesForUser(ctx, user.ID)
	if err != nil {
		return fmt.Errorf("courses: %w", err)
	}

	narrative := usecase.BuildNarrative(user.FirstName, user.LastName, interests, courses, user.Bio)
	vec, err := embedder.EmbedText(ctx, narrative)
	if err != nil {
		return fmt.Errorf("embed: %w", err)
	}
	if len(vec) == 0 {
		return fmt.Errorf("embed: empty vector")
	}

	if err := repo.UpdateEmbedding(ctx, user.ID, vec); err != nil {
		return fmt.Errorf("persist: %w", err)
	}
	return nil
}
