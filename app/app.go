package app

import (
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"

	"skeleton-go/app/controller"
	"skeleton-go/app/routes"
	"skeleton-go/internal/repository"
	"skeleton-go/internal/sessionstore"
	"skeleton-go/internal/usecase"
	"skeleton-go/pkg/cache"
	"skeleton-go/pkg/config"
	"skeleton-go/pkg/database"
	"skeleton-go/pkg/external/jsonplaceholder"
	"skeleton-go/pkg/i18n"
	"skeleton-go/pkg/logger"
	"skeleton-go/pkg/middleware"
	"skeleton-go/pkg/session"
	"skeleton-go/pkg/token"
)

// Application wires dependencies together and runs the Fiber HTTP server.
type Application struct {
	fiber *fiber.App
	cfg   *config.Config
	ti18n *i18n.Translator
}

// NewApplication builds the dependency graph for the HTTP stack.
func NewApplication() *Application {
	cfg := config.Load()

	fiberApp := fiber.New()

	logOpts := logger.Options{
		File:       cfg.Logging.File,
		MaxSizeMB:  cfg.Logging.MaxSizeMB,
		MaxBackups: cfg.Logging.MaxBackups,
		MaxAgeDays: cfg.Logging.MaxAgeDays,
		Compress:   cfg.Logging.Compress,
	}
	log, err := logger.New(logOpts)
	if err != nil {
		panic(err)
	}

	translator, err := i18n.LoadTranslator("resources/i18n", cfg.App.Locale)
	if err != nil {
		panic(err)
	}

	accessTTL := time.Duration(cfg.Auth.AccessTokenTTLMinutes) * time.Minute
	if accessTTL <= 0 {
		accessTTL = time.Hour
	}
	refreshTTL := time.Duration(cfg.Auth.RefreshTokenTTLMinutes) * time.Minute
	if refreshTTL <= 0 {
		refreshTTL = 24 * time.Hour
	}
	tokenManager := token.NewManager(cfg.App.Secret, accessTTL)
	authMiddleware := middleware.Authenticate(tokenManager, log)

	fiberApp.Use(recover.New())
	fiberApp.Use(cors.New())
	fiberApp.Use(translator.Middleware())
	fiberApp.Use(session.Middleware())
	fiberApp.Use(logger.HTTPLogger(log))

	db, err := database.NewMySQL(cfg.Database)
	if err != nil {
		panic(err)
	}

	gormDB, err := database.NewGORM(db, log, cfg.Database.LogQueries)
	if err != nil {
		panic(err)
	}
	userRepository := repository.NewGormUserRepository(gormDB, log)
	redisClient, err := cache.NewRedis(cfg.Redis)
	if err != nil {
		panic(err)
	}
	sessionStore := sessionstore.NewRedisRefreshTokenStore(redisClient)
	userUsecase := usecase.NewUserUsecase(userRepository)
	userUsecaseV2 := usecase.NewUserUsecaseV2(userRepository)
	authUsecase := usecase.NewAuthUsecase(userRepository, tokenManager, sessionStore, refreshTTL)
	authController := controller.NewAuthController(authUsecase, log, translator)
	placeholderClient := jsonplaceholder.NewClient(cfg.External.JSONPlaceholderURL, nil)
	externalController := controller.NewExternalController(placeholderClient, log, translator)
	userControllers := make([]controller.UserHTTPController, 0, len(cfg.App.Versions))
	for _, version := range cfg.App.Versions {
		switch strings.ToLower(version) {
		case "v2":
			userControllers = append(userControllers, controller.NewUserControllerV2(userUsecaseV2, version, log, translator))
		default:
			userControllers = append(userControllers, controller.NewUserController(userUsecase, version, log, translator))
		}
	}

	routes.RegisterAPIRoutes(fiberApp, cfg, translator, authController, authMiddleware, externalController, userControllers...)

	return &Application{
		fiber: fiberApp,
		cfg:   cfg,
		ti18n: translator,
	}
}

// Start runs the Fiber application using the configured port.
func (a *Application) Start() error {
	address := fmt.Sprintf(":%s", a.cfg.HTTP.Port)
	return a.fiber.Listen(address)
}
