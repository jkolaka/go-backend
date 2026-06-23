package main

import (
	"go-backend/config"
	"go-backend/handlers"
	"github.com/gin-gonic/gin"
)

func main() {
	config.InitDB()

	router := gin.Default()
	router.POST("/recipes", handlers.NewRecipeHandler)
	router.GET("/recipes", handlers.ListRecipesHandler)
	router.GET("/recipes/search", handlers.SearchRecipesHandler)
	router.PUT("/recipes/:id", handlers.UpdateRecipeHandler)
	router.DELETE("/recipes/:id", handlers.DeleteRecipeHandler)
	router.Run()
}
