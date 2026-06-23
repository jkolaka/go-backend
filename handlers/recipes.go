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

// Helper function to invalidate cache when data updates
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

	recipe.ID = xid.New().String()
	recipe.PublishedAt = time.Now()

	if err := config.DB.Create(&recipe).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create recipe"})
		return
	}

	clearCache() 
	c.JSON(http.StatusOK, recipe)
}

func ListRecipesHandler(c *gin.Context) {
	//Attempt to fetch data from Redis Cache
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

	// Fetch from SQLite Database when cache is not found
	var recipes []models.Recipe
	if err := config.DB.Find(&recipes).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch recipes"})
		return
	}

	// Saves to Redis with a 10-minute expiration
	if config.RedisClient != nil {
		if data, err := json.Marshal(recipes); err == nil {
			config.RedisClient.Set(config.Ctx, "recipes", string(data), 10*time.Minute)
		}
	}

	c.JSON(http.StatusOK, recipes)
}

func UpdateRecipeHandler(c *gin.Context) {
	id := c.Param("id")
	var existingRecipe models.Recipe

	if err := config.DB.First(&existingRecipe, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Recipe not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		}
		return
	}

	var updateData models.Recipe
	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updateData.ID = existingRecipe.ID
	updateData.PublishedAt = existingRecipe.PublishedAt

	if err := config.DB.Save(&updateData).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update recipe"})
		return
	}

	clearCache() 
	c.JSON(http.StatusOK, updateData)
}

func DeleteRecipeHandler(c *gin.Context) {
	id := c.Param("id")
	result := config.DB.Delete(&models.Recipe{}, "id = ?", id)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete recipe"})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Recipe not found"})
		return
	}

	clearCache() 
	c.JSON(http.StatusOK, gin.H{"message": "Recipe has been deleted"})
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