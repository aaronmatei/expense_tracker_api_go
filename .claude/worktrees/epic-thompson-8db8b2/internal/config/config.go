package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	AppEnv           string   `envconfig:"APP_ENV" default:"development"`
	Port             string   `envconfig:"PORT" default:"8080"`
	LogLevel         string   `envconfig:"LOG_LEVEL" default:"info"`
	DatabaseURL      string   `envconfig:"DATABASE_URL" required:"true"`
	JWTSecret        string   `envconfig:"JWT_SECRET" required:"true"`
	JWTExpiryMinutes int      `envconfig:"JWT_EXPIRY_MINUTES" default:"60"`
	CORSOrigins      []string `envconfig:"CORS_ORIGINS" default:"http://localhost:3000"`
}

func (c *Config) JWTExpiry() time.Duration {
	return time.Duration(c.JWTExpiryMinutes) * time.Minute
}

func (c *Config) IsProduction() bool {
	return strings.EqualFold(c.AppEnv, "production")
}

func Load() (*Config, error) {
	// godotenv.Load() reads the .env file (if it exists) and loads the variables into the environment. --- IGNORE ---
	_ = godotenv.Load()

	var cfg Config

	if err := envconfig.Process("", &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config from environment: %w", err)
	}

	return &cfg, nil
}
