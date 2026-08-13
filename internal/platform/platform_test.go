package platform_test

import (
	"errors"
	"testing"

	"github.com/walfa-labs/backend/internal/config"
	"github.com/walfa-labs/backend/internal/domain"
	"github.com/walfa-labs/backend/internal/platform"
)

type sampleRequest struct {
	Name  string `json:"name" validate:"required"`
	Email string `json:"email" validate:"required,email"`
	Role  string `json:"role" validate:"oneof=admin editor viewer"`
	Age   int    `json:"age" validate:"min=18,max=99"`
}

func TestPlatformValidator(t *testing.T) {
	logger := platform.NewLogger("development")
	v := platform.NewValidator()
	sv := platform.NewStructValidator(v, logger)

	t.Run("valid struct passes validation", func(t *testing.T) {
		req := sampleRequest{
			Name:  "Walfa",
			Email: "walfa@example.com",
			Role:  "admin",
			Age:   25,
		}

		err := sv.Validate(req)
		if err != nil {
			t.Fatalf("expected nil error for valid struct, got %v", err)
		}
	})

	t.Run("invalid struct returns domain.ValidationError with field details", func(t *testing.T) {
		req := sampleRequest{
			Name:  "",
			Email: "invalid-email",
			Role:  "superuser", // not in oneof
			Age:   10,          // less than 18
		}

		err := sv.Validate(req)
		if err == nil {
			t.Fatal("expected validation error for invalid struct, got nil")
		}

		var valErr *domain.ValidationError
		if !errors.As(err, &valErr) {
			t.Fatalf("expected *domain.ValidationError, got %T: %v", err, err)
		}

		if len(valErr.Fields) != 4 {
			t.Fatalf("expected 4 field errors, got %d", len(valErr.Fields))
		}

		fieldMap := make(map[string]string)
		for _, f := range valErr.Fields {
			fieldMap[f.Field] = f.Issue
		}

		if fieldMap["Name"] != "is required" {
			t.Errorf("unexpected issue for Name: %s", fieldMap["Name"])
		}
		if fieldMap["Email"] != "must be a valid email address" {
			t.Errorf("unexpected issue for Email: %s", fieldMap["Email"])
		}
		if fieldMap["Role"] != "must be one of: admin, editor, viewer" {
			t.Errorf("unexpected issue for Role: %s", fieldMap["Role"])
		}
		if fieldMap["Age"] != "must be at least 18 characters" {
			t.Errorf("unexpected issue for Age: %s", fieldMap["Age"])
		}
	})

	t.Run("non-struct input returns original error", func(t *testing.T) {
		err := sv.Validate("not-a-struct")
		if err == nil {
			t.Fatal("expected error validating non-struct, got nil")
		}
		var valErr *domain.ValidationError
		if errors.As(err, &valErr) {
			t.Errorf("did not expect *domain.ValidationError for non-struct input")
		}
	})
}

func TestPlatformJSON(t *testing.T) {
	fcfg := platform.FiberConfig()
	if fcfg.AppName != "Portfolio API v1" {
		t.Errorf("expected AppName 'Portfolio API v1', got '%s'", fcfg.AppName)
	}

	type payload struct {
		Message string `json:"message"`
		Count   int    `json:"count"`
	}

	data := payload{Message: "hello", Count: 42}
	bytes, err := fcfg.JSONEncoder(data)
	if err != nil {
		t.Fatalf("JSONEncoder error: %v", err)
	}

	var decoded payload
	if err := fcfg.JSONDecoder(bytes, &decoded); err != nil {
		t.Fatalf("JSONDecoder error: %v", err)
	}

	if decoded.Message != "hello" || decoded.Count != 42 {
		t.Errorf("unexpected decoded values: %+v", decoded)
	}
}

func TestPlatformServer(t *testing.T) {
	logger := platform.NewLogger("development")
	cfg := &config.Config{
		AppEnv:  "development",
		AppPort: ":8080",
	}

	app := platform.NewServer(cfg, logger)
	if app == nil {
		t.Fatal("expected non-nil Fiber app")
	}
}
