package routes

import (
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"

	"skeleton-go/app/controller"
	"skeleton-go/pkg/config"
	"skeleton-go/pkg/i18n"
	"skeleton-go/pkg/response"
)

const userIDRoute = "/users/:id"

// RegisterAPIRoutes wires HTTP endpoints to controllers.
func RegisterAPIRoutes(
	app *fiber.App,
	cfg *config.Config,
	translator *i18n.Translator,
	authController *controller.AuthController,
	authMiddleware fiber.Handler,
	externalController *controller.ExternalController,
	userControllers ...controller.UserHTTPController,
) {
	app.Get("/health", func(c *fiber.Ctx) error {
		locale := translator.LocaleFromContext(c)
		return response.Success(c, fiber.StatusOK, translator.T(locale, "health.ok"), fiber.Map{
			"name":     cfg.App.Name,
			"versions": cfg.App.Versions,
		})
	})

	api := app.Group("/api")
	for _, userController := range userControllers {
		version := userController.Version()
		versionGroup := api.Group(fmt.Sprintf("/%s", version))

		authGroup := versionGroup.Group("/auth")
		authGroup.Post("/login", authController.Login)
		authGroup.Post("/refresh", authController.Refresh)
		protected := versionGroup.Group("", authMiddleware)
		protected.Post("/auth/logout", authController.Logout)

		if externalController != nil && strings.EqualFold(version, "v1") {
			protected.Get("/placeholder/posts/:id", externalController.GetPlaceholderPost)
		}

		protected.Get("/users", userController.List)
		protected.Get(userIDRoute, userController.Get)
		protected.Post("/users", userController.Create)
		protected.Put(userIDRoute, userController.Update)
		protected.Delete(userIDRoute, userController.Delete)
	}
}
