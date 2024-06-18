package app

import (
	grpcapp "auth/internal/app/grpc"
	"auth/internal/config"
	"auth/internal/services"
	"auth/internal/storage/mongodb"
	"log/slog"
)

type App struct {
	GRPCServer *grpcapp.App
}

func New(log *slog.Logger, cfg *config.Config) *App {
	storage, err := mongodb.New(cfg.Storage.Path, cfg.Storage.Database, cfg.Storage.Collection)
	if err != nil {
		panic(err)
	}

	authService := services.New(log, storage)

	grpcApp := grpcapp.New(log, cfg, authService)

	return &App{
		GRPCServer: grpcApp,
	}
}
