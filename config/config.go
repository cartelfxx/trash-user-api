package config

import (
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Server     ServerConfig
	Discord    DiscordConfig
	Cache      CacheConfig
	RateLimit  RateLimitConfig
	Logging    LoggingConfig
}

type ServerConfig struct {
	Host         string
	Port         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

type DiscordConfig struct {
	Token         string
	APIURL        string
	APIVersion    string
	RequestTimeout time.Duration
	MaxRetries    int
	RetryDelay    time.Duration
}

type CacheConfig struct {
	Enabled           bool
	TTL               time.Duration
	MaxSize           int
	CleanupInterval   time.Duration
	AutoRefresh       bool
	RefreshInterval   time.Duration
}

type RateLimitConfig struct {
	Enabled           bool
	RequestsPerMinute int
	BurstSize         int
}

type LoggingConfig struct {
	Level      string
	Format     string
	WithEmojis bool
}

func Load() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		return nil, err
	}

	config := &Config{
		Server: ServerConfig{
			Host:         getEnv("HOST", "0.0.0.0"),
			Port:         getEnv("PORT", "8080"),
			ReadTimeout:  getDurationEnv("READ_TIMEOUT", 30*time.Second),
			WriteTimeout: getDurationEnv("WRITE_TIMEOUT", 30*time.Second),
			IdleTimeout:  getDurationEnv("IDLE_TIMEOUT", 60*time.Second),
		},
		Discord: DiscordConfig{
			Token:         getEnv("DISCORD_TOKEN", ""),
			APIURL:        getEnv("DISCORD_API_URL", "https://discord.com/api"),
			APIVersion:    getEnv("DISCORD_API_VERSION", "v9"),
			RequestTimeout: getDurationEnv("DISCORD_REQUEST_TIMEOUT", 30*time.Second),
			MaxRetries:    getIntEnv("DISCORD_MAX_RETRIES", 3),
			RetryDelay:    getDurationEnv("DISCORD_RETRY_DELAY", 1*time.Second),
		},
		Cache: CacheConfig{
			Enabled:         getBoolEnv("CACHE_ENABLED", true),
			TTL:             getDurationEnv("CACHE_TTL", 5*time.Minute),
			MaxSize:         getIntEnv("CACHE_MAX_SIZE", 1000),
			CleanupInterval: getDurationEnv("CACHE_CLEANUP_INTERVAL", 10*time.Minute),
			AutoRefresh:     getBoolEnv("CACHE_AUTO_REFRESH", true),
			RefreshInterval: getDurationEnv("CACHE_REFRESH_INTERVAL", 2*time.Minute),
		},
		RateLimit: RateLimitConfig{
			Enabled:           getBoolEnv("RATE_LIMIT_ENABLED", true),
			RequestsPerMinute: getIntEnv("RATE_LIMIT_REQUESTS_PER_MINUTE", 60),
			BurstSize:         getIntEnv("RATE_LIMIT_BURST_SIZE", 10),
		},
		Logging: LoggingConfig{
			Level:      getEnv("LOG_LEVEL", "info"),
			Format:     getEnv("LOG_FORMAT", "text"),
			WithEmojis: getBoolEnv("LOG_WITH_EMOJIS", true),
		},
	}

	return config, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getIntEnv(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getBoolEnv(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
} 