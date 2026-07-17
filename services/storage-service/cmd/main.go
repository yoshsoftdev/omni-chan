package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/omnichannel/omni-chan-v2/services/storage-service/internal/config"
	"github.com/omnichannel/omni-chan-v2/services/storage-service/internal/handler"
	"github.com/omnichannel/omni-chan-v2/services/storage-service/internal/middleware"
	"github.com/omnichannel/omni-chan-v2/services/storage-service/internal/storage"
	"github.com/omnichannel/omni-chan-v2/services/storage-service/pkg/db"
	"github.com/omnichannel/omni-chan-v2/services/storage-service/pkg/jwt"
	"github.com/omnichannel/omni-chan-v2/services/storage-service/pkg/migrate"
	"github.com/omnichannel/omni-chan-v2/services/storage-service/pkg/response"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}

	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	if err := migrate.Up(cfg.Database.URL, "migrations"); err != nil {
		log.Error("migration failed", "error", err)
		os.Exit(1)
	}

	database, err := db.New(ctx, cfg.Database)
	if err != nil {
		log.Error("database connection failed", "error", err)
		os.Exit(1)
	}
	defer database.Close()

	publicBaseURL := cfg.Storage.PublicBaseURL
	if publicBaseURL == "" {
		publicBaseURL = fmt.Sprintf("http://127.0.0.1:%d", cfg.HTTP.Port)
	}

	storageRepo := storage.NewRepository(database)
	store, err := storage.New(ctx, cfg.Storage, publicBaseURL, storageRepo)
	if err != nil {
		log.Error("storage init failed", "error", err)
		os.Exit(1)
	}

	jwtManager := jwt.NewManager(cfg.JWT.AccessSecret, cfg.JWT.RefreshSecret, cfg.JWT.AccessTTL)
	storageHandler := handler.NewStorageHandler(store)

	app := fiber.New(fiber.Config{
		AppName:      cfg.App.AppName + "-storage",
		ReadTimeout:  cfg.HTTP.ReadTimeout,
		WriteTimeout: cfg.HTTP.WriteTimeout,
		BodyLimit:    100 * 1024 * 1024,
	})
	app.Use(recover.New())
	app.Use(requestid.New())

	api := app.Group("/api/v1")
	api.Get("/health", func(c *fiber.Ctx) error {
		return response.OK(c, "ok", fiber.Map{"service": "storage-service", "driver": store.Driver()})
	})

	protected := api.Group("/storage", middleware.Auth(jwtManager, cfg.InternalToken))
	protected.Post("/upload", storageHandler.Upload)
	protected.Get("/blob/*", storageHandler.GetObject)
	protected.Get("/objects/:id/content", storageHandler.GetObjectContent)
	protected.Get("/objects/:id", storageHandler.GetMetadata)
	protected.Delete("/objects/:id", storageHandler.Delete)

	go func() {
		log.Info("storage service starting", "addr", cfg.HTTP.Addr(), "driver", store.Driver())
		if err := app.Listen(cfg.HTTP.Addr()); err != nil {
			log.Error("server stopped", "error", err)
		}
	}()

	<-ctx.Done()
	_ = app.Shutdown()
}
