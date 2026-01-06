package session

import (
	"context"
	"crypto/rand"
	"encoding/hex"

	"github.com/gofiber/fiber/v2"
)

type contextKey struct{}

// Session tracks contextual metadata for logging and tracing.
type Session struct {
	ID string
}

// New allocates a fresh session with a unique identifier.
func New() *Session {
	return &Session{ID: newID()}
}

// FromContext extracts the session from the supplied context, if any.
func FromContext(ctx context.Context) *Session {
	if ctx == nil {
		return nil
	}

	if sess, ok := ctx.Value(contextKey{}).(*Session); ok {
		return sess
	}

	return nil
}

// ContextWithSession embeds the session inside the provided context.
func ContextWithSession(ctx context.Context, sess *Session) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}

	return context.WithValue(ctx, contextKey{}, sess)
}

// Middleware attaches a session to each incoming HTTP request.
func Middleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		sess := New()
		ctx := ContextWithSession(c.UserContext(), sess)
		c.SetUserContext(ctx)
		c.Set("X-Request-ID", sess.ID)
		return c.Next()
	}
}

var randRead = rand.Read

func newID() string {
	var buf [8]byte
	if _, err := randRead(buf[:]); err != nil {
		return "sess-unknown"
	}
	return hex.EncodeToString(buf[:])
}
