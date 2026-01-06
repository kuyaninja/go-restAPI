package middleware

import (
	"context"
	"strings"

	"github.com/gofiber/fiber/v2"

	"skeleton-go/pkg/logger"
	"skeleton-go/pkg/token"
)

type userIDContextKey struct{}

// Authenticate validates bearer tokens and injects the user id into the request context.
func Authenticate(manager *token.Manager, log *logger.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		tokenStr := extractToken(c)
		if tokenStr == "" {
			return fiber.NewError(fiber.StatusUnauthorized, "missing token")
		}

		claims, err := manager.Parse(tokenStr)
		if err != nil {
			if log != nil {
				log.Error(c.UserContext(), "token parse failed", err, logger.Fields{"token": "invalid"})
			}
			return fiber.NewError(fiber.StatusUnauthorized, "invalid token")
		}

		ctx := context.WithValue(c.UserContext(), userIDContextKey{}, claims.UserID)
		c.SetUserContext(ctx)
		c.Locals("user_id", claims.UserID)

		return c.Next()
	}
}

// UserIDFromContext retrieves the authenticated user id.
func UserIDFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if id, ok := ctx.Value(userIDContextKey{}).(string); ok {
		return id
	}
	return ""
}

func extractToken(c *fiber.Ctx) string {
	header := c.Get("Authorization")
	if header != "" {
		parts := strings.SplitN(header, " ", 2)
		if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
			return strings.TrimSpace(parts[1])
		}
	}
	if token := c.Get("X-Auth-Token"); token != "" {
		return token
	}
	return ""
}
