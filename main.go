package main

import (
	"log"
	"pr-review-service/internal/handlers"
	"pr-review-service/internal/service"
	"pr-review-service/internal/storage"

	"github.com/gin-gonic/gin"
)

func main() {
	// Инициализация хранилища
	store, err := storage.NewPostgresStorage()
	if err != nil {
		log.Fatal("Failed to initialize storage:", err)
	}
	defer store.Close()

	// Инициализация сервиса
	reviewService := service.NewReviewService(store)

	// Инициализация обработчиков
	handler := handlers.NewHandler(reviewService)

	// Настройка роутера
	r := gin.Default()

	// Teams
	r.POST("/team/add", handler.CreateTeam)
	r.GET("/team/get", handler.GetTeam)

	// Users
	r.POST("/users/setIsActive", handler.SetUserActive)
	r.GET("/users/getReview", handler.GetUserReviews)

	// Pull Requests
	r.POST("/pullRequest/create", handler.CreatePR)
	r.POST("/pullRequest/merge", handler.MergePR)
	r.POST("/pullRequest/reassign", handler.ReassignReviewer)

	// Health
	r.GET("/health", handler.HealthCheck)

	log.Println("Server starting on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
