package controller

import (
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
	notFoundMessageKeyV2      = "users.not_found"
	invalidUserIDMessageKeyV2 = "users.invalid_id"
	invalidUserIDLogFieldV2   = "invalid user id"
)

// UserControllerV2 exposes HTTP handlers for version 2 of the user resource.
type UserControllerV2 struct {
	usecase usecase.UserUsecaseV2
	version string
	logger  *logger.Logger
	trans   *i18n.Translator
}

// NewUserControllerV2 creates a controller that serves the v2 API.
func NewUserControllerV2(usecase usecase.UserUsecaseV2, version string, log *logger.Logger, translator *i18n.Translator) *UserControllerV2 {
	return &UserControllerV2{usecase: usecase, version: version, logger: log, trans: translator}
}

func (uc *UserControllerV2) logError(c *fiber.Ctx, message string, err error, fields logger.Fields) {
	if uc.logger == nil {
		return
	}
	uc.logger.Error(c.UserContext(), message, err, fields)
}

// Version returns the API version the controller serves.
func (uc *UserControllerV2) Version() string {
	return uc.version
}

// List handles GET /users and returns every user.
func (uc *UserControllerV2) List(c *fiber.Ctx) error {
	locale := uc.locale(c)
	users, err := uc.usecase.ListUsers(c.UserContext())
	if err != nil {
		uc.logError(c, "user list failed", err, logger.Fields{"version": uc.version})
		return response.Failure(c, http.StatusInternalServerError, uc.trans.T(locale, "users.fetch_failed"), err.Error())
	}

	return response.Success(c, http.StatusOK, uc.trans.T(locale, "users.list.success"), fiber.Map{"version": uc.version, "items": users})
}

// Get handles GET /users/:id.
func (uc *UserControllerV2) Get(c *fiber.Ctx) error {
	locale := uc.locale(c)
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		uc.logError(c, invalidUserIDLogFieldV2, err, logger.Fields{"version": uc.version})
		return response.Failure(c, http.StatusBadRequest, uc.trans.T(locale, invalidUserIDMessageKeyV2), err.Error())
	}

	user, err := uc.usecase.GetUser(c.UserContext(), id)
	if err != nil {
		status := http.StatusInternalServerError
		if err == repository.ErrUserNotFound {
			status = http.StatusNotFound
		}
		uc.logError(c, "user fetch failed", err, logger.Fields{"version": uc.version, "user_id": id.String()})
		msgKey := "users.fetch_failed"
		if err == repository.ErrUserNotFound {
			msgKey = notFoundMessageKeyV2
		}
		return response.Failure(c, status, uc.trans.T(locale, msgKey), err.Error())
	}

	return response.Success(c, http.StatusOK, uc.trans.T(locale, "users.get.success"), fiber.Map{"version": uc.version, "user": user})
}

// Create handles POST /users.
func (uc *UserControllerV2) Create(c *fiber.Ctx) error {
	locale := uc.locale(c)
	var input usecase.CreateUserV2Input
	if err := c.BodyParser(&input); err != nil {
		uc.logError(c, "user create payload invalid", err, logger.Fields{"version": uc.version})
		return response.Failure(c, http.StatusBadRequest, uc.trans.T(locale, "users.payload_invalid"), err.Error())
	}

	user, err := uc.usecase.CreateUser(c.UserContext(), input)
	if err != nil {
		uc.logError(c, "user create failed", err, logger.Fields{"version": uc.version})
		return response.Failure(c, http.StatusBadRequest, uc.trans.T(locale, "users.create_failed"), err.Error())
	}

	return response.Success(c, http.StatusCreated, uc.trans.T(locale, "users.create.success"), fiber.Map{"version": uc.version, "user": user})
}

// Update handles PUT /users/:id.
func (uc *UserControllerV2) Update(c *fiber.Ctx) error {
	locale := uc.locale(c)
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		uc.logError(c, invalidUserIDLogFieldV2, err, logger.Fields{"version": uc.version})
		return response.Failure(c, http.StatusBadRequest, uc.trans.T(locale, invalidUserIDMessageKeyV2), err.Error())
	}

	var input usecase.UpdateUserV2Input
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
			msgKey = notFoundMessageKeyV2
		}
		return response.Failure(c, status, uc.trans.T(locale, msgKey), err.Error())
	}

	return response.Success(c, http.StatusOK, uc.trans.T(locale, "users.update.success"), fiber.Map{"version": uc.version, "user": user})
}

// Delete handles DELETE /users/:id.
func (uc *UserControllerV2) Delete(c *fiber.Ctx) error {
	locale := uc.locale(c)
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		uc.logError(c, invalidUserIDLogFieldV2, err, logger.Fields{"version": uc.version})
		return response.Failure(c, http.StatusBadRequest, uc.trans.T(locale, invalidUserIDMessageKeyV2), err.Error())
	}

	if err := uc.usecase.DeleteUser(c.UserContext(), id); err != nil {
		status := http.StatusInternalServerError
		if err == repository.ErrUserNotFound {
			status = http.StatusNotFound
		}
		uc.logError(c, "user delete failed", err, logger.Fields{"version": uc.version, "user_id": id.String()})
		msgKey := "users.delete_failed"
		if err == repository.ErrUserNotFound {
			msgKey = notFoundMessageKeyV2
		}
		return response.Failure(c, status, uc.trans.T(locale, msgKey), err.Error())
	}

	return response.Success(c, http.StatusOK, uc.trans.T(locale, "users.delete.success"), fiber.Map{"version": uc.version})
}

func (uc *UserControllerV2) locale(c *fiber.Ctx) string {
	if uc.trans == nil {
		return ""
	}
	return uc.trans.LocaleFromContext(c)
}
