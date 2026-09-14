package config

import (
	"os"
	"time"
)

type Config struct {
	Port      string
	MySQLDSN  string
	JWTSecret []byte
	TokenTTL  time.Duration
}

func Load() Config {
	return Config{
		Port:      getEnv("PORT", "8080"),
		MySQLDSN:  getEnv("MYSQL_DSN", "root@tcp(127.0.0.1:3306)/devops_camp?parseTime=true&charset=utf8mb4"),
		JWTSecret: []byte(getEnv("JWT_SECRET", "dev-secret-change-me")),
		TokenTTL:  time.Hour,
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
