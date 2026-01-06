package controller

import (
	"errors"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"skeleton-go/internal/repository"
	"skeleton-go/internal/usecase"
	"skeleton-go/pkg/i18n"
	"skeleton-go/pkg/logger"
	"skeleton-go/pkg/response"
)

const (
	notFoundMessageKey      = "users.not_found"
	invalidUserIDMessageKey = "users.invalid_id"
	invalidUserIDLogField   = "invalid user id"
)

// UserController exposes HTTP handlers for the user resource.
type UserController struct {
	usecase usecase.UserUsecase
	version string
	logger  *logger.Logger
	trans   *i18n.Translator
}

// NewUserController creates a new controller instance.
func NewUserController(usecase usecase.UserUsecase, version string, log *logger.Logger, translator *i18n.Translator) *UserController {
	return &UserController{usecase: usecase, version: version, logger: log, trans: translator}
}

func (uc *UserController) logError(ctx *fiber.Ctx, message string, err error, fields logger.Fields) {
	if uc.logger == nil {
		return
	}
	uc.logger.Error(ctx.UserContext(), message, err, fields)
}

// Version returns the API version the controller serves.
func (uc *UserController) Version() string {
	return uc.version
}

// List handles GET /users and returns every user.
func (uc *UserController) List(c *fiber.Ctx) error {
	locale := uc.locale(c)
	users, err := uc.usecase.ListUsers(c.UserContext())
	if err != nil {
		uc.logError(c, "user list failed", err, logger.Fields{"version": uc.version})
		return response.Failure(c, http.StatusInternalServerError, uc.trans.T(locale, "users.fetch_failed"), err.Error())
	}

	return response.Success(c, http.StatusOK, uc.trans.T(locale, "users.list.success"), users)
}

// Get handles GET /users/:id.
func (uc *UserController) Get(c *fiber.Ctx) error {
	locale := uc.locale(c)
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		uc.logError(c, invalidUserIDLogField, err, logger.Fields{"version": uc.version})
		return response.Failure(c, http.StatusBadRequest, uc.trans.T(locale, invalidUserIDMessageKey), err.Error())
	}

	user, err := uc.usecase.GetUser(c.UserContext(), id)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, repository.ErrUserNotFound) {
			status = http.StatusNotFound
		}
		uc.logError(c, "user fetch failed", err, logger.Fields{"version": uc.version, "user_id": id.String()})
		msgKey := "users.fetch_failed"
		if errors.Is(err, repository.ErrUserNotFound) {
			msgKey = notFoundMessageKey
		}
		return response.Failure(c, status, uc.trans.T(locale, msgKey), err.Error())
	}

	return response.Success(c, http.StatusOK, uc.trans.T(locale, "users.get.success"), user)
}

// Create handles POST /users.
func (uc *UserController) Create(c *fiber.Ctx) error {
	locale := uc.locale(c)
	var input usecase.CreateUserInput
	if err := c.BodyParser(&input); err != nil {
		uc.logError(c, "user create payload invalid", err, logger.Fields{"version": uc.version})
		return response.Failure(c, http.StatusBadRequest, uc.trans.T(locale, "users.payload_invalid"), err.Error())
	}

	user, err := uc.usecase.CreateUser(c.UserContext(), input)
	if err != nil {
		uc.logError(c, "user create failed", err, logger.Fields{"version": uc.version})
		return response.Failure(c, http.StatusBadRequest, uc.trans.T(locale, "users.create_failed"), err.Error())
	}

	return response.Success(c, http.StatusCreated, uc.trans.T(locale, "users.create.success"), user)
}

// Update handles PUT /users/:id.
func (uc *UserController) Update(c *fiber.Ctx) error {
	locale := uc.locale(c)
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		uc.logError(c, invalidUserIDLogField, err, logger.Fields{"version": uc.version})
		return response.Failure(c, http.StatusBadRequest, uc.trans.T(locale, invalidUserIDMessageKey), err.Error())
	}

	var input usecase.UpdateUserInput
	if err := c.BodyParser(&input); err != nil {
		uc.logError(c, "user update payload invalid", err, logger.Fields{"version": uc.version, "user_id": id.String()})
		return response.Failure(c, http.StatusBadRequest, uc.trans.T(locale, "users.payload_invalid"), err.Error())
	}

	user, err := uc.usecase.UpdateUser(c.UserContext(), id, input)
	if err != nil {
		status := http.StatusInternalServerError
		if err == repository.ErrUserNotFound {
			status = http.StatusNotFound
		}
		uc.logError(c, "user update failed", err, logger.Fields{"version": uc.version, "user_id": id.String()})
		msgKey := "users.update_failed"
		if err == repository.ErrUserNotFound {
			msgKey = notFoundMessageKey
		}
		return response.Failure(c, status, uc.trans.T(locale, msgKey), err.Error())
	}

	return response.Success(c, http.StatusOK, uc.trans.T(locale, "users.update.success"), user)
}

// Delete handles DELETE /users/:id.
func (uc *UserController) Delete(c *fiber.Ctx) error {
	locale := uc.locale(c)
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		uc.logError(c, invalidUserIDLogField, err, logger.Fields{"version": uc.version})
		return response.Failure(c, http.StatusBadRequest, uc.trans.T(locale, invalidUserIDMessageKey), err.Error())
	}

	if err := uc.usecase.DeleteUser(c.UserContext(), id); err != nil {
		status := http.StatusInternalServerError
		if err == repository.ErrUserNotFound {
			status = http.StatusNotFound
		}
		uc.logError(c, "user delete failed", err, logger.Fields{"version": uc.version, "user_id": id.String()})
		msgKey := "users.delete_failed"
		if err == repository.ErrUserNotFound {
			msgKey = notFoundMessageKey
		}
		return response.Failure(c, status, uc.trans.T(locale, msgKey), err.Error())
	}

	return response.Success(c, http.StatusOK, uc.trans.T(locale, "users.delete.success"), nil)
}

func (uc *UserController) locale(c *fiber.Ctx) string {
	if uc.trans == nil {
		return ""
	}
	return uc.trans.LocaleFromContext(c)
}
