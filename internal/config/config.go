package config

import (
	"errors"
	"os"
	"strings"
	"time"
)

const defaultJWTSecret = "dev-secret-change-me"

// ローカル開発時、Next.jsの開発サーバーはポート3000が使用中だと自動で3001などに
// ずれることがあるため、よく使われる番号をいくつかデフォルトで許可しておく。
const defaultCORSAllowedOrigins = "http://localhost:3000,http://localhost:3001"

type Config struct {
	Port               string
	MySQLDSN           string
	JWTSecret          []byte
	TokenTTL           time.Duration
	CORSAllowedOrigins []string
}

func Load() (Config, error) {
	jwtSecret := getEnv("JWT_SECRET", defaultJWTSecret)
	if getEnv("APP_ENV", "") == "production" && jwtSecret == defaultJWTSecret {
		return Config{}, errors.New("JWT_SECRET must be set explicitly when APP_ENV=production")
	}

	return Config{
		Port:               getEnv("PORT", "8080"),
		MySQLDSN:           getEnv("MYSQL_DSN", "root@tcp(127.0.0.1:3306)/devops_camp?parseTime=true&charset=utf8mb4"),
		JWTSecret:          []byte(jwtSecret),
		TokenTTL:           7 * 24 * time.Hour,
		CORSAllowedOrigins: splitCSV(getEnv("CORS_ALLOWED_ORIGINS", defaultCORSAllowedOrigins)),
	}, nil
}

func splitCSV(s string) []string {
	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
