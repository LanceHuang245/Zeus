package config

import (
	"Zephyr/internal/models"
	"context"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/joho/godotenv"
)

var (
	RedisClient    *redis.Client
	Ctx            = context.Background()
	RedisAddr      string
	RedisPassword  string
	RedisDB        int
	CacheTTL       time.Duration
	QweatherConfig models.QweatherConfig
	QweatherUrl    string

	// Server configuration
	ServerPort string
	EnableTLS  bool
	CertFile   string
	KeyFile    string
)

// LoadConfig loads configuration from .env file
func LoadConfig() {
	// Load .env file
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: .env file not found, using environment variables or defaults")
	}

	// Redis configuration
	RedisAddr = getEnv("REDIS_ADDR", "127.0.0.1:6379")
	RedisPassword = getEnv("REDIS_PASSWORD", "")
	RedisDB = getEnvInt("REDIS_DB", 0)

	// Cache TTL (default 30 minutes)
	cacheTTLMinutes := getEnvInt("CACHE_TTL_MINUTES", 30)
	CacheTTL = time.Duration(cacheTTLMinutes) * time.Minute

	// QWeather configuration
	QweatherConfig = models.QweatherConfig{
		ProjectID:     getEnv("QWEATHER_PROJECT_ID", ""),
		KeyID:         getEnv("QWEATHER_KEY_ID", ""),
		PrivateKeyPem: getEnv("QWEATHER_PRIVATE_KEY", ""),
	}

	QweatherUrl = getEnv("QWEATHER_URL", "")

	// Server configuration
	ServerPort = getEnv("SERVER_PORT", ":3899")
	EnableTLS = getEnvBool("ENABLE_TLS", true)
	CertFile = getEnv("CERT_FILE", "./cert/zephyr.claret.space_bundle.crt")
	KeyFile = getEnv("KEY_FILE", "./cert/zephyr.claret.space.key")
}

// getEnv gets environment variable with default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvInt gets environment variable as integer with default value
func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// getEnvBool gets environment variable as boolean with default value
func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

// InitRedis initializes Redis client
func InitRedis() {
	RedisClient = redis.NewClient(&redis.Options{
		Addr:     RedisAddr,
		Password: RedisPassword,
		DB:       RedisDB,
	})

	// Test Redis connection
	_, err := RedisClient.Ping(Ctx).Result()
	if err != nil {
		log.Printf("Failed to connect to Redis: %v", err)
	} else {
		log.Println("Successfully connected to Redis")
	}
}
