package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	AppEnv     string
	AppPort    string
	JwtSecret  string
	JwtExpires time.Duration
	AdminUser  string
	AdminPass  string
	DBHost     string
	DBPort     string
	DBUser     string
	DBPass     string
	DBName     string
	DBParams   string
}

func LoadFromEnv() Config {
	cfg := Config{
		AppEnv:    getEnv("APP_ENV", "dev"),
		AppPort:   getEnv("APP_PORT", "8080"),
		JwtSecret: getEnv("JWT_SECRET", "dev-secret"),
		AdminUser: getEnv("ADMIN_USER", "admin"),
		AdminPass: getEnv("ADMIN_PASS", "admin123"),
		DBHost:    getEnv("DB_HOST", ""),
		DBPort:    getEnv("DB_PORT", "3306"),
		DBUser:    getEnv("DB_USER", ""),
		DBPass:    getEnv("DB_PASS", ""),
		DBName:    getEnv("DB_NAME", ""),
		DBParams:  getEnv("DB_PARAMS", "charset=utf8mb4&parseTime=true&loc=Local"),
	}

	if expires, err := strconv.Atoi(getEnv("JWT_EXPIRES", "7200")); err == nil {
		cfg.JwtExpires = time.Duration(expires) * time.Second
	} else {
		cfg.JwtExpires = 2 * time.Hour
	}

	return cfg
}

// DSN returns a MySQL DSN string if DBName is set; otherwise returns empty.
func (c Config) DSN() string {
	if c.DBName == "" {
		return ""
	}
	host := c.DBHost
	if host == "" {
		host = "127.0.0.1"
	}
	port := c.DBPort
	if port == "" {
		port = "3306"
	}
	// mysql DSN: user:pass@tcp(host:port)/dbname?params
	return c.DBUser + ":" + c.DBPass + "@tcp(" + host + ":" + port + ")/" + c.DBName + "?" + c.DBParams
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
