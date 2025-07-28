package storageapp

import (
	gorm "diaryhub/sso-service/internal/storage/gorm"
	_ "diaryhub/sso-service/internal/storage/postgresql"
	"fmt"

	"log/slog"
)

type App struct {
	log         *slog.Logger
	storagePath string
	Storage     *gorm.Storage
}

func MustConnect(logger *slog.Logger, storagePath string) *App {
	app, err := Connect(logger, storagePath)
	if err != nil {
		panic(err)
	}
	return app
}

func Connect(logger *slog.Logger, storagePath string) (*App, error) {
	const op = "app.storage.Connect"
	log := logger.With(slog.String("op", op))

	log.Info("Connecting to storage...")
	strg, err := gorm.New(storagePath)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	log.Info("Storage connected")
	return &App{log: log, storagePath: storagePath, Storage: strg}, nil
}

func (a *App) Disconnect() {
	const op = "app.storage.Stop"
	log := a.log.With(slog.String("op", op))

	log.Info("STOPPING storage")
	a.Storage.Close()
}
