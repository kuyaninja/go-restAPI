package controller

import (
	"context"
	"net/http"
	"strings"

	"github.com/gofiber/fiber/v2"

	"skeleton-go/pkg/external/jsonplaceholder"
	"skeleton-go/pkg/i18n"
	"skeleton-go/pkg/logger"
	"skeleton-go/pkg/response"
)

// PlaceholderClient describes the JSONPlaceholder client behavior we rely on.
type PlaceholderClient interface {
	GetPost(ctx context.Context, id string) (jsonplaceholder.Post, error)
}

// ExternalController exposes endpoints that aggregate third-party resources.
type ExternalController struct {
	placeholder PlaceholderClient
	logger      *logger.Logger
	trans       *i18n.Translator
}

// NewExternalController builds a controller that proxies JSONPlaceholder.
func NewExternalController(client PlaceholderClient, log *logger.Logger, translator *i18n.Translator) *ExternalController {
	return &ExternalController{placeholder: client, logger: log, trans: translator}
}

// GetPlaceholderPost fetches a sample post from JSONPlaceholder.
func (c *ExternalController) GetPlaceholderPost(ctx *fiber.Ctx) error {
	locale := c.locale(ctx)
	id := strings.TrimSpace(ctx.Params("id"))
	if id == "" {
		return response.Failure(ctx, http.StatusBadRequest, c.translated(locale, "external.post.invalid_id"), "post id is required")
	}

	post, err := c.placeholder.GetPost(ctx.UserContext(), id)
	if err != nil {
		c.logError(ctx, "placeholder post fetch failed", err, logger.Fields{"post_id": id})
		return response.Failure(ctx, http.StatusBadGateway, c.translated(locale, "external.post.fetch_failed"), err.Error())
	}

	return response.Success(ctx, http.StatusOK, c.translated(locale, "external.post.fetch_success"), post)
}

func (c *ExternalController) locale(ctx *fiber.Ctx) string {
	if c.trans == nil {
		return ""
	}
	return c.trans.LocaleFromContext(ctx)
}

func (c *ExternalController) translated(locale, key string) string {
	if c.trans == nil {
		return key
	}
	return c.trans.T(locale, key)
}

func (c *ExternalController) logError(ctx *fiber.Ctx, message string, err error, fields logger.Fields) {
	if c.logger == nil {
		return
	}
	c.logger.Error(ctx.UserContext(), message, err, fields)
}
