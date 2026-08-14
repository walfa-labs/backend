package config_test

import (
	"os"
	"testing"
	"time"

	"github.com/walfa-labs/backend/internal/config"
)

func setRequiredEnvs() {
	os.Setenv("DB_USER", "portfolio")
	os.Setenv("DB_PASSWORD", "pass")
	os.Setenv("DB_NAME", "FREEPDB1")
	os.Setenv("JWT_SECRET", "super-secret-jwt-key")
	os.Setenv("ADMIN_PASSWORD_HASH", "$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy")
	os.Setenv("OCI_TENANCY_OCID", "ocid1.tenancy.oc1..test")
	os.Setenv("OCI_USER_OCID", "ocid1.user.oc1..test")
	os.Setenv("OCI_FINGERPRINT", "00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00")
	os.Setenv("OCI_REGION", "ap-singapore-1")
	os.Setenv("OCI_PRIVATE_KEY_PATH", "./test_key.pem")
	os.Setenv("OCI_BUCKET", "test-bucket")
}

func clearEnvs() {
	vars := []string{
		"APP_ENV", "APP_PORT", "DB_DRIVER", "DB_HOST", "DB_PORT",
		"DB_USER", "DB_PASSWORD", "DB_NAME", "DB_SSLMODE",
		"ANALYTICS_DB_HOST", "ANALYTICS_DB_PORT", "ANALYTICS_DB_USER",
		"ANALYTICS_DB_PASSWORD", "ANALYTICS_DB_NAME",
		"MIGRATE_URL", "MIGRATE_ANALYTICS_URL",
		"JWT_SECRET", "JWT_ACCESS_TTL", "JWT_REFRESH_TTL",
		"ADMIN_USERNAME", "ADMIN_PASSWORD_HASH",
		"STORAGE_DRIVER", "STORAGE_LOCAL_DIR", "STORAGE_BASE_URL",
		"OCI_TENANCY_OCID", "OCI_USER_OCID", "OCI_FINGERPRINT",
		"OCI_REGION", "OCI_PRIVATE_KEY_PATH", "OCI_NAMESPACE",
		"OCI_BUCKET", "CORS_ALLOWED_ORIGINS",
	}
	for _, v := range vars {
		os.Unsetenv(v)
	}
}

func TestConfigLoad(t *testing.T) {
	t.Run("fails when required envs are missing", func(t *testing.T) {
		clearEnvs()
		cfg, err := config.Load()
		if err == nil {
			t.Fatalf("expected error when required env vars are missing, got cfg: %+v", cfg)
		}
	})

	t.Run("succeeds with defaults when required envs are provided", func(t *testing.T) {
		clearEnvs()
		setRequiredEnvs()
		defer clearEnvs()

		cfg, err := config.Load()
		if err != nil {
			t.Fatalf("unexpected error loading config: %v", err)
		}

		if cfg.AppEnv != "development" {
			t.Errorf("expected default AppEnv 'development', got '%s'", cfg.AppEnv)
		}
		if cfg.AppPort != ":8080" {
			t.Errorf("expected default AppPort ':8080', got '%s'", cfg.AppPort)
		}
		if cfg.DBDriver != "oracle" {
			t.Errorf("expected default DBDriver 'oracle', got '%s'", cfg.DBDriver)
		}
		if cfg.DBHost != "localhost" {
			t.Errorf("expected default DBHost 'localhost', got '%s'", cfg.DBHost)
		}
		if cfg.DBPort != "1521" {
			t.Errorf("expected default DBPort '1521', got '%s'", cfg.DBPort)
		}
		if cfg.DBUser != "portfolio" {
			t.Errorf("expected DBUser 'portfolio', got '%s'", cfg.DBUser)
		}
		if cfg.DBPassword != "pass" {
			t.Errorf("expected DBPassword 'pass', got '%s'", cfg.DBPassword)
		}
		if cfg.DBName != "FREEPDB1" {
			t.Errorf("expected DBName 'FREEPDB1', got '%s'", cfg.DBName)
		}
		if cfg.DBSSLMode != "disable" {
			t.Errorf("expected default DBSSLMode 'disable', got '%s'", cfg.DBSSLMode)
		}
		if cfg.OracleDSN() != "portfolio/pass@localhost:1521/FREEPDB1" {
			t.Errorf("unexpected OracleDSN: %s", cfg.OracleDSN())
		}
		if cfg.OracleAnalyticsDSN() != "portfolio/pass@localhost:1521/FREEPDB1" {
			t.Errorf("unexpected OracleAnalyticsDSN fallback: %s", cfg.OracleAnalyticsDSN())
		}
		if cfg.PostgresDSN() != "postgres://portfolio:pass@localhost:1521/FREEPDB1?sslmode=disable" {
			t.Errorf("unexpected PostgresDSN: %s", cfg.PostgresDSN())
		}
		if cfg.StorageDriver != "local" {
			t.Errorf("expected default StorageDriver 'local', got '%s'", cfg.StorageDriver)
		}
		if cfg.LocalStorageDir != "./uploads" {
			t.Errorf("expected default LocalStorageDir './uploads', got '%s'", cfg.LocalStorageDir)
		}
		if cfg.LocalStorageBaseURL != "http://localhost:8080/uploads" {
			t.Errorf("expected default LocalStorageBaseURL 'http://localhost:8080/uploads', got '%s'", cfg.LocalStorageBaseURL)
		}
		if cfg.JWTAccessTTL != 15*time.Minute {
			t.Errorf("expected default JWTAccessTTL 15m, got %v", cfg.JWTAccessTTL)
		}
		if cfg.JWTRefreshTTL != 168*time.Hour {
			t.Errorf("expected default JWTRefreshTTL 168h, got %v", cfg.JWTRefreshTTL)
		}
		if cfg.AdminUsername != "admin" {
			t.Errorf("expected default AdminUsername 'admin', got '%s'", cfg.AdminUsername)
		}
		if len(cfg.CORSAllowedOrigins) != 1 || cfg.CORSAllowedOrigins[0] != "http://localhost:3000" {
			t.Errorf("unexpected CORSAllowedOrigins: %v", cfg.CORSAllowedOrigins)
		}

		if !cfg.IsDevelopment() {
			t.Errorf("expected IsDevelopment() to be true")
		}
		if cfg.IsProduction() {
			t.Errorf("expected IsProduction() to be false")
		}
		if cfg.JWTSecret() != "super-secret-jwt-key" {
			t.Errorf("expected JWTSecret() to return 'super-secret-jwt-key', got '%s'", cfg.JWTSecret())
		}
	})

	t.Run("parses custom overrides correctly", func(t *testing.T) {
		clearEnvs()
		setRequiredEnvs()
		defer clearEnvs()

		os.Setenv("APP_ENV", "production")
		os.Setenv("APP_PORT", ":9090")
		os.Setenv("DB_DRIVER", "postgres")
		os.Setenv("DB_HOST", "pg-host")
		os.Setenv("DB_PORT", "5432")
		os.Setenv("DB_USER", "pguser")
		os.Setenv("DB_PASSWORD", "pgpass")
		os.Setenv("DB_NAME", "pgdb")
		os.Setenv("DB_SSLMODE", "require")
		os.Setenv("JWT_ACCESS_TTL", "30m")
		os.Setenv("JWT_REFRESH_TTL", "72h")
		os.Setenv("CORS_ALLOWED_ORIGINS", "https://walfa.dev,https://admin.walfa.dev")

		cfg, err := config.Load()
		if err != nil {
			t.Fatalf("unexpected error loading config: %v", err)
		}

		if !cfg.IsProduction() {
			t.Errorf("expected IsProduction() to be true")
		}
		if cfg.IsDevelopment() {
			t.Errorf("expected IsDevelopment() to be false")
		}
		if cfg.AppPort != ":9090" {
			t.Errorf("expected AppPort ':9090', got '%s'", cfg.AppPort)
		}
		if cfg.DBDriver != "postgres" {
			t.Errorf("expected DBDriver 'postgres', got '%s'", cfg.DBDriver)
		}
		if cfg.PostgresDSN() != "postgres://pguser:pgpass@pg-host:5432/pgdb?sslmode=require" {
			t.Errorf("unexpected PostgresDSN: %s", cfg.PostgresDSN())
		}
		if cfg.JWTAccessTTL != 30*time.Minute {
			t.Errorf("expected JWTAccessTTL 30m, got %v", cfg.JWTAccessTTL)
		}
		if len(cfg.CORSAllowedOrigins) != 2 || cfg.CORSAllowedOrigins[0] != "https://walfa.dev" || cfg.CORSAllowedOrigins[1] != "https://admin.walfa.dev" {
			t.Errorf("unexpected CORSAllowedOrigins: %v", cfg.CORSAllowedOrigins)
		}
	})

	t.Run("builds OracleAnalyticsDSN with custom analytics parameters", func(t *testing.T) {
		clearEnvs()
		setRequiredEnvs()
		defer clearEnvs()

		os.Setenv("ANALYTICS_DB_HOST", "adw-host")
		os.Setenv("ANALYTICS_DB_PORT", "1522")
		os.Setenv("ANALYTICS_DB_USER", "adwuser")
		os.Setenv("ANALYTICS_DB_PASSWORD", "adwpass")
		os.Setenv("ANALYTICS_DB_NAME", "ADWPDB")

		cfg, err := config.Load()
		if err != nil {
			t.Fatalf("unexpected error loading config: %v", err)
		}

		expectedDSN := "adwuser/adwpass@adw-host:1522/ADWPDB"
		if cfg.OracleAnalyticsDSN() != expectedDSN {
			t.Errorf("expected OracleAnalyticsDSN '%s', got '%s'", expectedDSN, cfg.OracleAnalyticsDSN())
		}
	})

	t.Run("fails when invalid DB_DRIVER is provided", func(t *testing.T) {
		clearEnvs()
		setRequiredEnvs()
		defer clearEnvs()

		os.Setenv("DB_DRIVER", "mysql")

		_, err := config.Load()
		if err == nil {
			t.Fatal("expected error for invalid DB_DRIVER, got nil")
		}
	})

	t.Run("fails when STORAGE_DRIVER is oci but OCI credentials are missing", func(t *testing.T) {
		clearEnvs()
		setRequiredEnvs()
		defer clearEnvs()

		os.Setenv("STORAGE_DRIVER", "oci")
		os.Unsetenv("OCI_TENANCY_OCID")

		_, err := config.Load()
		if err == nil {
			t.Fatal("expected error when STORAGE_DRIVER=oci with missing credentials, got nil")
		}
	})
}
