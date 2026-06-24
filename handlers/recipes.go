package handlers

import (
	"encoding/json"
	"errors"
	"go-backend/config"
	"go-backend/models"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/xid"
	"gorm.io/gorm"
)

func clearCache() {
	if config.RedisClient != nil {
		config.RedisClient.Del(config.Ctx, "recipes")
	}
}

func NewRecipeHandler(c *gin.Context) {
	var recipe models.Recipe
	if err := c.ShouldBindJSON(&recipe); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	recipe.ID = xid.New().String()
	recipe.PublishedAt = time.Now()
	recipe.UserID = userID.(string)

	if err := config.DB.Create(&recipe).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create recipe"})
		return
	}

	clearCache()
	c.JSON(http.StatusCreated, recipe)
}

func ListRecipesHandler(c *gin.Context) {
	// Try Redis cache
	if config.RedisClient != nil {
		val, err := config.RedisClient.Get(config.Ctx, "recipes").Result()
		if err == nil {
			var recipes []models.Recipe
			if err := json.Unmarshal([]byte(val), &recipes); err == nil {
				c.JSON(http.StatusOK, recipes)
				return
			}
		}
	}

	// Cache miss - get from database
	var recipes []models.Recipe
	if err := config.DB.Preload("User").Find(&recipes).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch recipes"})
		return
	}

	// Store in cache
	if config.RedisClient != nil {
		if data, err := json.Marshal(recipes); err == nil {
			config.RedisClient.Set(config.Ctx, "recipes", data, 10*time.Minute)
		}
	}

	c.JSON(http.StatusOK, recipes)
}

func UpdateRecipeHandler(c *gin.Context) {
	id := c.Param("id")
	userID, _ := c.Get("userId")

	var existingRecipe models.Recipe
	if err := config.DB.First(&existingRecipe, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Recipe not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		}
		return
	}

	// Check ownership
	if existingRecipe.UserID != userID.(string) {
		c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to update this recipe"})
		return
	}

	var updateData models.Recipe
	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Preserve ID, timestamps, and UserID
	updateData.ID = existingRecipe.ID
	updateData.PublishedAt = existingRecipe.PublishedAt
	updateData.UserID = existingRecipe.UserID

	if err := config.DB.Save(&updateData).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update recipe"})
		return
	}

	clearCache()
	c.JSON(http.StatusOK, updateData)
}

func DeleteRecipeHandler(c *gin.Context) {
	id := c.Param("id")
	userID, _ := c.Get("userId")

	var existingRecipe models.Recipe
	if err := config.DB.First(&existingRecipe, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Recipe not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		}
		return
	}

	// Check ownership
	if existingRecipe.UserID != userID.(string) {
		c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to delete this recipe"})
		return
	}

	result := config.DB.Delete(&models.Recipe{}, "id = ?", id)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete recipe"})
		return
	}

	clearCache()
	c.JSON(http.StatusOK, gin.H{"message": "Recipe deleted successfully"})
}

func SearchRecipesHandler(c *gin.Context) {
	tag := c.Query("tag")
	if tag == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Tag parameter is required"})
		return
	}

	var recipes []models.Recipe
	query := config.DB.Where("LOWER(tags) LIKE ?", "%"+strings.ToLower(tag)+"%")
	if err := query.Find(&recipes).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to search recipes"})
		return
	}

	c.JSON(http.StatusOK, recipes)
}