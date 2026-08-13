package config_test

import (
	"os"
	"testing"
	"time"

	"github.com/walfa-labs/backend/internal/config"
)

func setRequiredEnvs() {
	os.Setenv("ATP_DSN", "portfolio/pass@atp_high")
	os.Setenv("ADW_DSN", "portfolio/pass@adw_high")
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
		"APP_ENV", "APP_PORT", "ATP_DSN", "ADW_DSN", "JWT_SECRET",
		"JWT_ACCESS_TTL", "JWT_REFRESH_TTL", "ADMIN_USERNAME",
		"ADMIN_PASSWORD_HASH", "OCI_TENANCY_OCID", "OCI_USER_OCID",
		"OCI_FINGERPRINT", "OCI_REGION", "OCI_PRIVATE_KEY_PATH",
		"OCI_NAMESPACE", "OCI_BUCKET", "CORS_ALLOWED_ORIGINS",
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
		if cfg.JWTAccessTTL != 30*time.Minute {
			t.Errorf("expected JWTAccessTTL 30m, got %v", cfg.JWTAccessTTL)
		}
		if len(cfg.CORSAllowedOrigins) != 2 || cfg.CORSAllowedOrigins[0] != "https://walfa.dev" || cfg.CORSAllowedOrigins[1] != "https://admin.walfa.dev" {
			t.Errorf("unexpected CORSAllowedOrigins: %v", cfg.CORSAllowedOrigins)
		}
	})
}
