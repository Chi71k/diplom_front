package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"

	"studybuddy/backend/pkg/db"
	"studybuddy/backend/pkg/embedding"
	"studybuddy/backend/services/groups/delivery"
	"studybuddy/backend/services/groups/repository"
	"studybuddy/backend/services/groups/usecase"
)

func main() {
	_ = godotenv.Load(".env")

	port := getEnv("GROUPS_SERVER_PORT", "8085")
	jwtSecret := getEnv("JWT_SECRET", "dev-secret-change-in-production")
	geminiKey := getEnv("GEMINI_API_KEY", "")
	pointsURL := getEnv("POINTS_SERVICE_URL", "")
	dsn := getEnv("DATABASE_URL", "postgres://studybuddy:studybuddy@localhost:5432/studybuddy?sslmode=disable")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := db.NewPool(ctx, dsn)
	if err != nil {
		log.Fatalf("failed to connect to db: %v", err)
	}
	defer pool.Close()

	groupRepo := repository.NewPgGroupRepository(pool)
	embCache := embedding.NewPgCache(pool)
	geminiClient := embedding.NewClient(geminiKey)
	embedProv := repository.NewEmbeddingProvider(geminiClient, embCache, pool)

	createUC := usecase.NewCreateGroup(groupRepo)
	getUC := usecase.NewGetGroup(groupRepo)
	listUC := usecase.NewListMyGroups(groupRepo)
	inviteUC := usecase.NewInviteMember(groupRepo, pointsURL, []byte(jwtSecret))
	removeUC := usecase.NewRemoveMember(groupRepo)
	deleteUC := usecase.NewDeleteGroup(groupRepo)
	updateUC := usecase.NewUpdateGroup(groupRepo)
	suggestUC := usecase.NewSuggestMembersForGroup(groupRepo, embedProv, geminiClient)

	handler := &delivery.GroupsHandler{
		Create:   createUC,
		Get:      getUC,
		ListMine: listUC,
		Invite:   inviteUC,
		Remove:   removeUC,
		Delete:   deleteUC,
		Update:   updateUC,
		Suggest:  suggestUC,
	}

	router := delivery.NewRouter(handler, []byte(jwtSecret))

	log.Printf("groups service listening on :%s", port)
	if err := http.ListenAndServe(":"+port, router); err != nil {
		log.Fatal(err)
	}
}

func getEnv(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}
