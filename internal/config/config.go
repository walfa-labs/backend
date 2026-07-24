package config

import (
	"fmt"
	"time"

	"github.com/caarlos0/env/v11"
)

// Config holds all application configuration loaded from environment variables.
type Config struct {
	AppEnv  string `env:"APP_ENV" envDefault:"development"`
	AppPort string `env:"APP_PORT" envDefault:":8080"`

	DatabaseURL string `env:"DATABASE_URL,required"`

	JWTSecretKey string        `env:"JWT_SECRET,required"`
	JWTAccessTTL time.Duration `env:"JWT_ACCESS_TTL" envDefault:"15m"`
	JWTRefreshTTL time.Duration `env:"JWT_REFRESH_TTL" envDefault:"168h"`

	AdminUsername     string `env:"ADMIN_USERNAME" envDefault:"admin"`
	AdminPasswordHash string `env:"ADMIN_PASSWORD_HASH,required"`

	ObjectStorage struct {
		Endpoint     string `env:"ENDPOINT" envDefault:"localhost:9000"`
		Bucket       string `env:"BUCKET" envDefault:"portfolio-assets"`
		AccessKey    string `env:"ACCESS_KEY" envDefault:"minio"`
		SecretKey    string `env:"SECRET_KEY" envDefault:"minio123"`
		UsePathStyle bool   `env:"USE_PATH_STYLE" envDefault:"true"`
	} `envPrefix:"OBJECT_STORAGE_"`

	CORSAllowedOrigins []string `env:"CORS_ALLOWED_ORIGINS" envDefault:"http://localhost:3000" envSeparator:","`
}

// Load reads configuration from environment variables.
func Load() (*Config, error) {
	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		return nil, fmt.Errorf("config: %w", err)
	}
	return &cfg, nil
}

// IsProduction returns true if the app is running in production mode.
func (c *Config) IsProduction() bool { return c.AppEnv == "production" }

// IsDevelopment returns true if the app is running in development mode.
func (c *Config) IsDevelopment() bool { return c.AppEnv == "development" || c.AppEnv == "" }

// JWTSecret returns the JWT signing secret. Satisfies middleware.jwtKeyReader.
func (c *Config) JWTSecret() string { return c.JWTSecretKey }
