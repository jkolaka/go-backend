package config

import (
	"context"
	"go-backend/models"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB
var RedisClient *redis.Client
var Ctx = context.Background()

// LoadEnv loads environment variables from .env file
func LoadEnv() {
	err := godotenv.Load()
	if err != nil {
		log.Println("⚠️  Warning: .env file not found, using system environment variables")
	}
}

// GetJWTSecret returns the JWT secret key from environment
func GetJWTSecret() []byte {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		log.Println("⚠️  Warning: JWT_SECRET not set, using default secret (INSECURE!)")
		return []byte("default-secret-key-change-me")
	}
	return []byte(secret)
}

// GetPort returns the port from environment
func GetPort() string {
	port := os.Getenv("PORT")
	if port == "" {
		return ":8080"
	}
	return ":" + port
}

// GetDatabaseURL returns the database URL from environment
func GetDatabaseURL() string {
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		return "./recipes.db"
	}
	return url
}

// GetRedisURL returns the Redis URL from environment
func GetRedisURL() string {
	url := os.Getenv("REDIS_URL")
	if url == "" {
		return "localhost:6379"
	}
	return url
}

func InitDB() {
	var err error
	dbURL := GetDatabaseURL()
	
	DB, err = gorm.Open(sqlite.Open(dbURL), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatal("Failed to connect to SQLite:", err)
	}

	// Auto migrate both models
	err = DB.AutoMigrate(&models.Recipe{}, &models.User{})
	if err != nil {
		log.Fatal("Failed to migrate SQLite schema:", err)
	}
}

func InitRedis() {
	redisURL := GetRedisURL()
	
	RedisClient = redis.NewClient(&redis.Options{
		Addr:     redisURL,
		Password: "",
		DB:       0,
	})

	_, err := RedisClient.Ping(Ctx).Result()
	if err != nil {
		log.Printf(" Warning: Could not connect to Redis at %s: %v\n", redisURL, err)
		RedisClient = nil 
	} 
}