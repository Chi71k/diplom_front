package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"studybuddy/backend/pkg/db"
	"studybuddy/backend/pkg/embedding"
	"studybuddy/backend/pkg/gemini"
	matchingrepo "studybuddy/backend/services/matching/repository"
	"studybuddy/backend/services/users/delivery"
	"studybuddy/backend/services/users/repository"
	"studybuddy/backend/services/users/usecase"

	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load(".env")
	port := getEnv("USERS_SERVER_PORT", "8081")
	jwtSecret := getEnv("JWT_SECRET", "dev-secret-change-in-production")
	geminiKey := getEnv("GEMINI_API_KEY", "")
	dsn := getEnv("DATABASE_URL", "postgres://studybuddy:studybuddy@localhost:5432/studybuddy?sslmode=disable")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := db.NewPool(ctx, dsn)
	if err != nil {
		log.Fatalf("failed to connect to db: %v", err)
	}
	defer pool.Close()

	profileRepo := repository.NewPgProfileRepository(pool)
	adminRepo := repository.NewPgAdminRepository(pool)
	interestRepo := repository.NewPgInterestRepository(pool)
	userIntrestRepo := repository.NewPgUserInterestRepository(pool)
	embCache := embedding.NewPgCache(pool)
	friendships := matchingrepo.NewPgFriendshipRepository(pool)

	var embedder gemini.Embedder = gemini.NoOpEmbedder{}
	if geminiKey != "" {
		e, err := gemini.NewEmbedder(geminiKey)
		if err != nil {
			log.Fatalf("gemini embedder: %v", err)
		}
		embedder = e
	} else {
		log.Print("GEMINI_API_KEY not set: student search and embedding generation are disabled")
	}

	embeddingRegenerator := usecase.NewEmbeddingRegenerator(profileRepo, embedder, profileRepo, embCache)
	searchStudentsUC := usecase.NewSearchStudents(profileRepo, embedder)

	getMeUC := usecase.NewGetMe(profileRepo)
	getUserUC := usecase.NewGetUser(profileRepo)
	updateMeUC := usecase.NewUpdateMe(profileRepo, embCache, embeddingRegenerator)
	deleteMeUC := usecase.NewDeleteMe(profileRepo)

	listInterestsUC := usecase.NewListInterests(interestRepo)
	getMyInterestsUC := usecase.NewGetMyInterests(userIntrestRepo)
	replaceMyInterestsUC := usecase.NewReplaceMyInterests(interestRepo, userIntrestRepo, embCache, embeddingRegenerator)

	listFriendsUC := usecase.NewListFriends(friendships)
	removeFriendUC := usecase.NewRemoveFriend(friendships)

	adminListUsersUC := usecase.NewAdminListUsers(adminRepo)
	adminSetUserActiveUC := usecase.NewAdminSetUserActive(adminRepo)
	adminStatsUC := usecase.NewAdminStats(adminRepo)

	usersHandler := &delivery.UsersHandler{
		GetMe:          getMeUC,
		GetUser:        getUserUC,
		UpdateMe:       updateMeUC,
		DeleteMe:       deleteMeUC,
		SearchStudents: searchStudentsUC,
	}
	intrestsHandler := &delivery.InterestsHandler{
		ListCatalog: listInterestsUC,
		GetMine:     getMyInterestsUC,
		ReplaceMine: replaceMyInterestsUC,
	}
	friendsHandler := &delivery.FriendsHandler{
		List:   listFriendsUC,
		Remove: removeFriendUC,
	}
	adminHandler := &delivery.AdminHandler{
		ListUsers:     adminListUsersUC,
		SetUserActive: adminSetUserActiveUC,
		Stats:         adminStatsUC,
	}

	router := delivery.NewRouter(usersHandler, intrestsHandler, friendsHandler, adminHandler, []byte(jwtSecret))

	log.Printf("users service listening on :%s", port)
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
