package session

import (
	"context"
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
)

func TestNewGeneratesID(t *testing.T) {
	sess := New()
	if sess.ID == "" {
		t.Fatalf("expected session id to be generated")
	}
	if len(sess.ID) != 16 {
		t.Fatalf("expected 16 hex chars, got %d", len(sess.ID))
	}
}

func TestContextWithSessionRoundTrip(t *testing.T) {
	sess := &Session{ID: "test"}
	ctx := ContextWithSession(context.Background(), sess)
	if got := FromContext(ctx); got != sess {
		t.Fatalf("expected same session pointer, got %#v", got)
	}

	if FromContext(nil) != nil {
		t.Fatalf("expected nil context to return nil session")
	}
	if val := FromContext(context.Background()); val != nil {
		t.Fatalf("expected empty context to have no session")
	}
}

func TestContextWithSessionHandlesNilContext(t *testing.T) {
	sess := &Session{ID: "abc"}
	ctx := ContextWithSession(nil, sess)
	if ctx == nil {
		t.Fatalf("expected non-nil context")
	}
	if got := FromContext(ctx); got != sess {
		t.Fatalf("session not stored: %#v", got)
	}
}

func TestMiddlewareAttachesSession(t *testing.T) {
	app := fiber.New()
	app.Use(Middleware())
	var sessionID string
	app.Get("/", func(c *fiber.Ctx) error {
		sess := FromContext(c.UserContext())
		if sess == nil || sess.ID == "" {
			return fiber.NewError(fiber.StatusInternalServerError, "missing session")
		}
		sessionID = sess.ID
		return nil
	})

	req := httptest.NewRequest("GET", "/", nil)
	resp, err := app.Test(req)
	if err != nil || resp.StatusCode != 200 {
		t.Fatalf("unexpected response: %v status %d", err, resp.StatusCode)
	}
	if sessionID == "" {
		t.Fatalf("session id not captured")
	}
	if resp.Header.Get("X-Request-ID") != sessionID {
		t.Fatalf("expected response header to match session id")
	}
}

func TestNewIDFallbackOnError(t *testing.T) {
	orig := randRead
	randRead = func([]byte) (int, error) { return 0, errors.New("boom") }
	defer func() { randRead = orig }()
	if id := newID(); id != "sess-unknown" {
		t.Fatalf("expected fallback id, got %q", id)
	}
}
