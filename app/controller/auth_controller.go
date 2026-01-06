package controller

import (
	"net/http"

	"github.com/gofiber/fiber/v2"

	"skeleton-go/internal/usecase"
	"skeleton-go/pkg/i18n"
	"skeleton-go/pkg/logger"
	"skeleton-go/pkg/response"
)

// AuthController exposes login endpoint.
type AuthController struct {
	authUsecase usecase.AuthUsecase
	logger      *logger.Logger
	translator  *i18n.Translator
}

// NewAuthController builds the auth controller.
func NewAuthController(authUsecase usecase.AuthUsecase, log *logger.Logger, translator *i18n.Translator) *AuthController {
	return &AuthController{authUsecase: authUsecase, logger: log, translator: translator}
}

// Login handles POST /auth/login requests.
func (ac *AuthController) Login(c *fiber.Ctx) error {
	locale := ac.locale(c)

	var input usecase.LoginInput
	if err := c.BodyParser(&input); err != nil {
		ac.logError(c, "login payload invalid", err, nil)
		return response.Failure(c, http.StatusBadRequest, ac.translator.T(locale, "auth.login.payload_invalid"), err.Error())
	}

	tokens, err := ac.authUsecase.Login(c.UserContext(), input)
	if err != nil {
		ac.logError(c, "login failed", err, logger.Fields{"email": input.Email})
		return response.Failure(c, http.StatusUnauthorized, ac.translator.T(locale, "auth.login.invalid"), err.Error())
	}

	return response.Success(c, http.StatusOK, ac.translator.T(locale, "auth.login.success"), tokensResponse(tokens))
}

// Refresh exchanges a refresh token for a new token pair.
func (ac *AuthController) Refresh(c *fiber.Ctx) error {
	locale := ac.locale(c)
	var input usecase.RefreshInput
	if err := c.BodyParser(&input); err != nil {
		ac.logError(c, "refresh payload invalid", err, nil)
		return response.Failure(c, http.StatusBadRequest, ac.translator.T(locale, "auth.refresh.payload_invalid"), err.Error())
	}

	tokens, err := ac.authUsecase.Refresh(c.UserContext(), input)
	if err != nil {
		ac.logError(c, "refresh failed", err, nil)
		return response.Failure(c, http.StatusUnauthorized, ac.translator.T(locale, "auth.refresh.invalid"), err.Error())
	}

	return response.Success(c, http.StatusOK, ac.translator.T(locale, "auth.refresh.success"), tokensResponse(tokens))
}

// Logout revokes the current user's refresh session.
func (ac *AuthController) Logout(c *fiber.Ctx) error {
	locale := ac.locale(c)
	userID := c.Locals("user_id")
	uid, _ := userID.(string)
	if uid == "" {
		return response.Failure(c, http.StatusUnauthorized, ac.translator.T(locale, "auth.logout.invalid"), "missing user id")
	}

	if err := ac.authUsecase.Logout(c.UserContext(), uid); err != nil {
		ac.logError(c, "logout failed", err, logger.Fields{"user_id": uid})
		return response.Failure(c, http.StatusInternalServerError, ac.translator.T(locale, "auth.logout.failed"), err.Error())
	}

	return response.Success(c, http.StatusOK, ac.translator.T(locale, "auth.logout.success"), nil)
}

func tokensResponse(tokens usecase.AuthTokens) fiber.Map {
	return fiber.Map{
		"access_token":  tokens.AccessToken,
		"refresh_token": tokens.RefreshToken,
		"expires_in":    tokens.ExpiresIn,
	}
}

func (ac *AuthController) logError(c *fiber.Ctx, message string, err error, fields logger.Fields) {
	if ac.logger != nil {
		ac.logger.Error(c.UserContext(), message, err, fields)
	}
}

func (ac *AuthController) locale(c *fiber.Ctx) string {
	if ac.translator == nil {
		return ""
	}
	return ac.translator.LocaleFromContext(c)
}
