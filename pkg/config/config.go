package config

import (
	"os"
	"strconv"
	"github.com/joho/godotenv"
)

type Config struct {
	App      AppConfig
	Database DatabaseConfig
	Redis    RedisConfig
	JWT      JWTConfig
	OAuth    OAuthConfig
	Bunny    BunnyConfig
}

type AppConfig struct {
	Name string
	Port string
	Env  string
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

type JWTConfig struct {
	Secret string
}

type OAuthConfig struct {
	// Google
	GoogleClientID     string
	GoogleClientSecret string
	GoogleRedirectURL  string

	// Facebook
	FacebookClientID     string
	FacebookClientSecret string
	FacebookRedirectURL  string

	// LINE
	LINEClientID     string
	LINEClientSecret string
	LINERedirectURL  string
}

type BunnyConfig struct {
	StorageZone string
	AccessKey   string
	BaseURL     string
	CDNUrl      string
}

func LoadConfig() (*Config, error) {
	err := godotenv.Load()
	if err != nil {
		return nil, err
	}

	redisDB, _ := strconv.Atoi(getEnv("REDIS_DB", "0"))

	config := &Config{
		App: AppConfig{
			Name: getEnv("APP_NAME", "GoFiber Template"),
			Port: getEnv("APP_PORT", "3000"),
			Env:  getEnv("APP_ENV", "development"),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", ""),
			DBName:   getEnv("DB_NAME", "gofiber_template"),
			SSLMode:  getEnv("DB_SSL_MODE", "disable"),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnv("REDIS_PORT", "6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       redisDB,
		},
		JWT: JWTConfig{
			Secret: getEnv("JWT_SECRET", "your-secret-key"),
		},
		OAuth: OAuthConfig{
			GoogleClientID:       getEnv("GOOGLE_CLIENT_ID", ""),
			GoogleClientSecret:   getEnv("GOOGLE_CLIENT_SECRET", ""),
			GoogleRedirectURL:    getEnv("GOOGLE_REDIRECT_URL", ""),
			FacebookClientID:     getEnv("FACEBOOK_CLIENT_ID", ""),
			FacebookClientSecret: getEnv("FACEBOOK_CLIENT_SECRET", ""),
			FacebookRedirectURL:  getEnv("FACEBOOK_REDIRECT_URL", ""),
			LINEClientID:         getEnv("LINE_CLIENT_ID", ""),
			LINEClientSecret:     getEnv("LINE_CLIENT_SECRET", ""),
			LINERedirectURL:      getEnv("LINE_REDIRECT_URL", ""),
		},
		Bunny: BunnyConfig{
			StorageZone: getEnv("BUNNY_STORAGE_ZONE", ""),
			AccessKey:   getEnv("BUNNY_ACCESS_KEY", ""),
			BaseURL:     getEnv("BUNNY_BASE_URL", "https://storage.bunnycdn.com"),
			CDNUrl:      getEnv("BUNNY_CDN_URL", ""),
		},
	}

	return config, nil
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}