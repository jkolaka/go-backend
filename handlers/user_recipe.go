package handlers

import (
	"go-backend/config"
	"go-backend/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetUserRecipesHandler(c *gin.Context) {
	userID, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var recipes []models.Recipe
	if err := config.DB.Where("user_id = ?", userID).Order("published_at DESC").Find(&recipes).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user recipes"})
		return
	}

	c.JSON(http.StatusOK, recipes)
}