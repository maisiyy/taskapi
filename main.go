package main

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"taskapi/db"
	"taskapi/handlers"
	"taskapi/middleware"
	"taskapi/models"
)

func main() {
	db.Connect()

	if err := models.CreateTable(); err != nil {
		log.Fatalf("❌ Failed to create tasks table: %v", err)
	}
	if err := models.CreateUsersTable(); err != nil {
		log.Fatalf("❌ Failed to create users table: %v", err)
	}
	fmt.Println("✅ Tables ready!")

	r := gin.Default()

	// Public routes — no token needed
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})
	r.POST("/register", handlers.Register)
	r.POST("/login", handlers.Login)

	// Protected routes — must have valid JWT
	protected := r.Group("/tasks")
	protected.Use(middleware.AuthRequired())
	{
		protected.GET("", handlers.GetAllTasks)
		protected.POST("", handlers.CreateTask)
		protected.GET("/:id", handlers.GetTask)
		protected.PUT("/:id", handlers.UpdateTask)
		protected.DELETE("/:id", handlers.DeleteTask)
	}

	fmt.Println("🚀 Server running on http://localhost:3000")
	r.Run(":3000")
}