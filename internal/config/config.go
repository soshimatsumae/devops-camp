package config

import (
	"errors"
	"os"
	"time"
)

const defaultJWTSecret = "dev-secret-change-me"

type Config struct {
	Port      string
	MySQLDSN  string
	JWTSecret []byte
	TokenTTL  time.Duration
}

func Load() (Config, error) {
	jwtSecret := getEnv("JWT_SECRET", defaultJWTSecret)
	if getEnv("APP_ENV", "") == "production" && jwtSecret == defaultJWTSecret {
		return Config{}, errors.New("JWT_SECRET must be set explicitly when APP_ENV=production")
	}

	return Config{
		Port:      getEnv("PORT", "8080"),
		MySQLDSN:  getEnv("MYSQL_DSN", "root@tcp(127.0.0.1:3306)/devops_camp?parseTime=true&charset=utf8mb4"),
		JWTSecret: []byte(jwtSecret),
		TokenTTL:  time.Hour,
	}, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
