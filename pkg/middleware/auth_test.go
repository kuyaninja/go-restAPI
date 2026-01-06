package middleware

import (
	"context"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"

	"skeleton-go/pkg/logger"
	"skeleton-go/pkg/token"
)

func TestAuthenticateRejectsMissingToken(t *testing.T) {
	app := fiber.New()
	manager := token.NewManager("secret", time.Minute)
	app.Use(Authenticate(manager, nil))
	app.Get("/secure", func(*fiber.Ctx) error { return nil })

	resp, err := app.Test(httptest.NewRequest("GET", "/secure", nil))
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}
}

func TestAuthenticateRejectsInvalidToken(t *testing.T) {
	app := fiber.New()
	manager := token.NewManager("secret", time.Minute)
	log, err := logger.New(logger.Options{})
	if err != nil {
		t.Fatalf("failed to build logger: %v", err)
	}
	app.Use(Authenticate(manager, log))
	app.Get("/secure", func(*fiber.Ctx) error { return nil })

	req := httptest.NewRequest("GET", "/secure", nil)
	req.Header.Set("Authorization", "Bearer invalid")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}
}

func TestAuthenticateSuccessSetsContext(t *testing.T) {
	app := fiber.New()
	manager := token.NewManager("secret", time.Minute)
	app.Use(Authenticate(manager, nil))
	app.Get("/secure", func(c *fiber.Ctx) error {
		if id := UserIDFromContext(c.UserContext()); id != "user-1" {
			t.Fatalf("expected user id in context, got %q", id)
		}
		if id := c.Locals("user_id"); id != "user-1" {
			t.Fatalf("expected user id in locals, got %v", id)
		}
		return c.SendStatus(fiber.StatusOK)
	})

	tokenStr, err := manager.Generate("user-1")
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	req := httptest.NewRequest("GET", "/secure", nil)
	req.Header.Set("X-Auth-Token", tokenStr)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

func TestUserIDFromContextHandlesMissingValues(t *testing.T) {
	if id := UserIDFromContext(nil); id != "" {
		t.Fatalf("expected empty id for nil context")
	}
	if id := UserIDFromContext(context.Background()); id != "" {
		t.Fatalf("expected empty id for missing value")
	}
	ctx := context.WithValue(context.Background(), userIDContextKey{}, "foo")
	if id := UserIDFromContext(ctx); id != "foo" {
		t.Fatalf("expected foo, got %q", id)
	}
}
