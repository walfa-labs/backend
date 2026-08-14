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

	// ATPDSN is the godror connect string for Autonomous Transaction
	// Processing (operational OLTP store), e.g. "user/password@dbname_high"
	// or "user/password@host:port/service_name" for local Oracle DB.
	ATPDSN string `env:"ATP_DSN,required"`
	// ADWDSN is the godror connect string for Autonomous Data Warehouse
	// (analytics store), e.g. "user/password@dbname_high"
	// or "user/password@host:port/service_name" for local Oracle DB.
	ADWDSN string `env:"ADW_DSN,required"`

	JWTSecretKey  string        `env:"JWT_SECRET,required"`
	JWTAccessTTL  time.Duration `env:"JWT_ACCESS_TTL" envDefault:"15m"`
	JWTRefreshTTL time.Duration `env:"JWT_REFRESH_TTL" envDefault:"168h"`

	AdminUsername     string `env:"ADMIN_USERNAME" envDefault:"admin"`
	AdminPasswordHash string `env:"ADMIN_PASSWORD_HASH,required"`

	// StorageDriver selects the asset store implementation: "local" (disk) or "oci" (OCI Object Storage).
	StorageDriver       string `env:"STORAGE_DRIVER" envDefault:"local"`
	LocalStorageDir     string `env:"STORAGE_LOCAL_DIR" envDefault:"./uploads"`
	LocalStorageBaseURL string `env:"STORAGE_BASE_URL" envDefault:"http://localhost:8080/uploads"`

	// OCI holds Oracle Cloud Infrastructure Object Storage credentials and
	// target bucket. Required only when STORAGE_DRIVER is "oci".
	OCI struct {
		TenancyOCID    string `env:"TENANCY_OCID"`
		UserOCID       string `env:"USER_OCID"`
		Fingerprint    string `env:"FINGERPRINT"`
		Region         string `env:"REGION"`
		PrivateKeyPath string `env:"PRIVATE_KEY_PATH"`
		Namespace      string `env:"NAMESPACE"`
		Bucket         string `env:"BUCKET"`
	} `envPrefix:"OCI_"`

	CORSAllowedOrigins []string `env:"CORS_ALLOWED_ORIGINS" envDefault:"http://localhost:3000" envSeparator:","`
}

// Load reads configuration from environment variables.
func Load() (*Config, error) {
	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		return nil, fmt.Errorf("config: %w", err)
	}

	if cfg.StorageDriver == "oci" {
		if cfg.OCI.TenancyOCID == "" || cfg.OCI.UserOCID == "" || cfg.OCI.Fingerprint == "" ||
			cfg.OCI.Region == "" || cfg.OCI.PrivateKeyPath == "" || cfg.OCI.Bucket == "" {
			return nil, fmt.Errorf("config: OCI credentials are required when STORAGE_DRIVER=oci")
		}
	}

	return &cfg, nil
}

// IsProduction returns true if the app is running in production mode.
func (c *Config) IsProduction() bool { return c.AppEnv == "production" }

// IsDevelopment returns true if the app is running in development mode.
func (c *Config) IsDevelopment() bool { return c.AppEnv == "development" || c.AppEnv == "" }

// JWTSecret returns the JWT signing secret. Satisfies middleware.jwtKeyReader.
func (c *Config) JWTSecret() string { return c.JWTSecretKey }
