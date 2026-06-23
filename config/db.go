package config

import (
	"context"
	"go-backend/models"
	"log"

	"github.com/redis/go-redis/v9"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB
var RedisClient *redis.Client
var Ctx = context.Background()

func InitDB() {
	var err error
	DB, err = gorm.Open(sqlite.Open("./recipes.db"), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to SQLite:", err)
	}

	err = DB.AutoMigrate(&models.Recipe{})
	if err != nil {
		log.Fatal("Failed to migrate SQLite schema:", err)
	}
}

func InitRedis() {
	RedisClient = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", 
		DB:       0,  
	})

	_, err := RedisClient.Ping(Ctx).Result()
	if err != nil {
		log.Println("Warning: Could not connect to Redis, continuing without cache layer. Error:", err)
	}
}