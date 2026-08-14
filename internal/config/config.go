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

	// DBDriver selects the database backend: "oracle" or "postgres".
	DBDriver string `env:"DB_DRIVER" envDefault:"oracle"`

	// Primary (OLTP) database connection parameters.
	DBHost     string `env:"DB_HOST" envDefault:"localhost"`
	DBPort     string `env:"DB_PORT" envDefault:"1521"`
	DBUser     string `env:"DB_USER,required"`
	DBPassword string `env:"DB_PASSWORD,required"`
	// DBName is the Oracle service name (e.g. "FREEPDB1") or PostgreSQL database name.
	DBName    string `env:"DB_NAME,required"`
	DBSSLMode string `env:"DB_SSLMODE" envDefault:"disable"`

	// Analytics (OLAP) database connection parameters — Oracle only.
	// When DB_DRIVER=postgres, the analytics store reuses the primary DB connection.
	// When DB_DRIVER=oracle and ANALYTICS_DB_HOST is empty, also reuses the primary connection.
	AnalyticsDBHost     string `env:"ANALYTICS_DB_HOST"`
	AnalyticsDBPort     string `env:"ANALYTICS_DB_PORT"`
	AnalyticsDBUser     string `env:"ANALYTICS_DB_USER"`
	AnalyticsDBPassword string `env:"ANALYTICS_DB_PASSWORD"`
	AnalyticsDBName     string `env:"ANALYTICS_DB_NAME"`

	// golang-migrate connection URLs (used only by the migrate-* Taskfile tasks).
	MigrateURL          string `env:"MIGRATE_URL"`
	MigrateAnalyticsURL string `env:"MIGRATE_ANALYTICS_URL"`

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

	switch cfg.DBDriver {
	case "oracle", "postgres":
		// valid
	default:
		return nil, fmt.Errorf("config: DB_DRIVER must be \"oracle\" or \"postgres\", got %q", cfg.DBDriver)
	}

	if cfg.StorageDriver == "oci" {
		if cfg.OCI.TenancyOCID == "" || cfg.OCI.UserOCID == "" || cfg.OCI.Fingerprint == "" ||
			cfg.OCI.Region == "" || cfg.OCI.PrivateKeyPath == "" || cfg.OCI.Bucket == "" {
			return nil, fmt.Errorf("config: OCI credentials are required when STORAGE_DRIVER=oci")
		}
	}

	return &cfg, nil
}

// OracleDSN builds a godror connect string: "user/password@host:port/service_name".
func (c *Config) OracleDSN() string {
	return fmt.Sprintf("%s/%s@%s:%s/%s", c.DBUser, c.DBPassword, c.DBHost, c.DBPort, c.DBName)
}

// OracleAnalyticsDSN builds the godror connect string for the analytics (ADW) database.
// Falls back to the primary DSN when ANALYTICS_DB_HOST is not set.
func (c *Config) OracleAnalyticsDSN() string {
	host := c.AnalyticsDBHost
	if host == "" {
		return c.OracleDSN()
	}
	port := c.AnalyticsDBPort
	if port == "" {
		port = c.DBPort
	}
	user := c.AnalyticsDBUser
	if user == "" {
		user = c.DBUser
	}
	pass := c.AnalyticsDBPassword
	if pass == "" {
		pass = c.DBPassword
	}
	name := c.AnalyticsDBName
	if name == "" {
		name = c.DBName
	}
	return fmt.Sprintf("%s/%s@%s:%s/%s", user, pass, host, port, name)
}

// PostgresDSN builds a PostgreSQL connection URL:
// "postgres://user:password@host:port/dbname?sslmode=mode".
func (c *Config) PostgresDSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		c.DBUser, c.DBPassword, c.DBHost, c.DBPort, c.DBName, c.DBSSLMode)
}

// IsProduction returns true if the app is running in production mode.
func (c *Config) IsProduction() bool { return c.AppEnv == "production" }

// IsDevelopment returns true if the app is running in development mode.
func (c *Config) IsDevelopment() bool { return c.AppEnv == "development" || c.AppEnv == "" }

// JWTSecret returns the JWT signing secret. Satisfies middleware.jwtKeyReader.
func (c *Config) JWTSecret() string { return c.JWTSecretKey }
