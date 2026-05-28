package config

import (
	"os"
)

// Config holds all configuration for the auth service
type Config struct {
	Server         ServerConfig
	Keycloak       KeycloakConfig
	Redis          RedisConfig
	Database       DatabaseConfig
	JWT            JWTConfig
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Port string
}

// KeycloakConfig holds Keycloak configuration
type KeycloakConfig struct {
	URL          string
	Realm        string
	ClientID     string
	ClientSecret string
	IssuerURL    string
	JWKSURL      string
}

// RedisConfig holds Redis configuration
type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

// DatabaseConfig holds database configuration for persistent session storage
type DatabaseConfig struct {
	Driver   string
	DSN      string
	MaxIdle  int
	MaxOpen  int
}

// JWTConfig holds JWT configuration
type JWTConfig struct {
	VerifySignature bool
	Audience        string
}

// Load loads configuration from environment variables
func Load() *Config {
	return &Config{
		Server: ServerConfig{
			Port: getEnv("AUTH_SERVICE_PORT", "8082"),
		},
		Keycloak: KeycloakConfig{
			URL:          getEnv("KEYCLOAK_URL", "http://keycloak:8080"),
			Realm:        getEnv("KEYCLOAK_REALM", "demo-realm"),
			ClientID:     getEnv("KEYCLOAK_CLIENT_ID", "spring-boot-app"),
			ClientSecret: getEnv("KEYCLOAK_CLIENT_SECRET", "spring-boot-secret-key-2024"),
			IssuerURL:    getEnv("KEYCLOAK_ISSUER_URL", ""),
			JWKSURL:      getEnv("KEYCLOAK_JWKS_URL", ""),
		},
		Redis: RedisConfig{
			Addr:     getEnv("REDIS_ADDR", "redis:6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       0,
		},
		Database: DatabaseConfig{
			Driver:  getEnv("DB_DRIVER", ""),
			DSN:     getEnv("DB_DSN", ""),
			MaxIdle: 5,
			MaxOpen: 20,
		},
		JWT: JWTConfig{
			VerifySignature: getEnv("JWT_VERIFY_SIGNATURE", "true") == "true",
			Audience:        getEnv("JWT_AUDIENCE", ""),
		},
	}
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
