package main

import (
	"log"
	"pr-review-service/internal/handlers"
	"pr-review-service/internal/service"
	"pr-review-service/internal/storage"

	"github.com/gin-gonic/gin"
)

func main() {
	store, err := storage.NewPostgresStorage()
	if err != nil {
		log.Fatal("Failed to initialize storage:", err)
	}
	defer store.Close()

	reviewService := service.NewReviewService(store)

	handler := handlers.NewHandler(reviewService)

	r := gin.Default()

	r.POST("/team/add", handler.CreateTeam)
	r.GET("/team/get", handler.GetTeam)

	r.POST("/users/setIsActive", handler.SetUserActive)
	r.GET("/users/getReview", handler.GetUserReviews)

	r.POST("/pullRequest/create", handler.CreatePR)
	r.POST("/pullRequest/merge", handler.MergePR)
	r.POST("/pullRequest/reassign", handler.ReassignReviewer)

	r.GET("/health", handler.HealthCheck)

	log.Println("Server starting on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
