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

	// Server settings
	ServerPort string
	EnableTLS  bool
	CertFile   string
	KeyFile    string
)

// Config is an immutable snapshot used when wiring application dependencies
// The package-level variables above remain temporarily for compatibility
// with legacy callers
type Config struct {
	RedisAddr     string
	RedisPassword string
	RedisDB       int
	CacheTTL      time.Duration

	QweatherConfig models.QweatherConfig
	QweatherURL    string

	OpenMeteoURL     string
	AirQualityURL    string
	OpenStreetMapURL string
	ServerPort       string
	EnableTLS        bool
	CertFile         string
	KeyFile          string
}

// Snapshot returns the current configuration as a dependency injection snapshot
func Snapshot() Config {
	return Config{
		RedisAddr:        RedisAddr,
		RedisPassword:    RedisPassword,
		RedisDB:          RedisDB,
		CacheTTL:         CacheTTL,
		QweatherConfig:   QweatherConfig,
		QweatherURL:      QweatherUrl,
		OpenMeteoURL:     OmForcastUrl,
		AirQualityURL:    OmAirQualityUrl,
		OpenStreetMapURL: OsmUrl,
		ServerPort:       ServerPort,
		EnableTLS:        EnableTLS,
		CertFile:         CertFile,
		KeyFile:          KeyFile,
	}
}

// LoadConfig loads configuration from the environment file
func LoadConfig() {
	// Load the environment file
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: .env file not found, using environment variables or defaults")
	}

	// Redis settings
	RedisAddr = getEnv("REDIS_ADDR", "127.0.0.1:26379")
	RedisPassword = getEnv("REDIS_PASSWORD", "RF6f7JecsbWFp8jP")
	RedisDB = getEnvInt("REDIS_DB", 0)

	// Cache lifetime with a default of 30 minutes
	cacheTTLMinutes := getEnvInt("CACHE_TTL_MINUTES", 30)
	CacheTTL = time.Duration(cacheTTLMinutes) * time.Minute

	// QWeather settings
	QweatherConfig = models.QweatherConfig{
		ProjectID:     getEnv("QWEATHER_PROJECT_ID", ""),
		KeyID:         getEnv("QWEATHER_KEY_ID", ""),
		PrivateKeyPem: getEnv("QWEATHER_PRIVATE_KEY", ""),
	}

	QweatherUrl = getEnv("QWEATHER_URL", "")

	// Server settings
	ServerPort = getEnv("SERVER_PORT", ":3899")
	EnableTLS = getEnvBool("ENABLE_TLS", true)
	CertFile = getEnv("CERT_FILE", "./cert/zephyr.claret.space_bundle.crt")
	KeyFile = getEnv("KEY_FILE", "./cert/zephyr.claret.space.key")
}

// getEnv reads an environment variable with a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvInt reads an integer environment variable with a default value
func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// getEnvBool reads a boolean environment variable with a default value
func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

// InitRedis initializes the Redis client
func InitRedis() {
	RedisClient = NewRedisClient(Snapshot())
}

// NewRedisClient creates a Redis adapter from an explicit configuration snapshot
// The connection check and fallback behavior match InitRedis
func NewRedisClient(cfg Config) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})

	// Test the Redis connection
	_, err := client.Ping(Ctx).Result()
	if err != nil {
		log.Printf("Failed to connect to Redis: %v", err)
	} else {
		log.Println("Successfully connected to Redis")
	}
	return client
}
