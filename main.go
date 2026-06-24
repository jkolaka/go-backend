package main

import (
	"go-backend/config"
	"go-backend/handlers"
	"go-backend/middleware"
	"github.com/gin-gonic/gin"
)

func main() {
	// Load environment variables
	config.LoadEnv()

	// Initialize database and Redis
	config.InitDB()
	config.InitRedis()

	router := gin.Default()

	// Public routes
	auth := router.Group("/auth")
	{
		auth.POST("/register", handlers.RegisterHandler)
		auth.POST("/login", handlers.LoginHandler)
	}

	router.GET("/recipes", handlers.ListRecipesHandler)
	router.GET("/recipes/search", handlers.SearchRecipesHandler)

	// Protected routes
	protected := router.Group("/")
	protected.Use(middleware.AuthMiddleware())
	{
		protected.GET("/profile", handlers.GetProfileHandler)
		protected.POST("/recipes", handlers.NewRecipeHandler)
		protected.PUT("/recipes/:id", handlers.UpdateRecipeHandler)
		protected.DELETE("/recipes/:id", handlers.DeleteRecipeHandler)
		protected.GET("/my-recipes", handlers.GetUserRecipesHandler)
	}

	// Start server
	port := config.GetPort()
	router.Run(port)
}