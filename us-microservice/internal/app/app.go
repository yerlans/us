package app

import (
	"log/slog"
	grpcapp "urlSh/internal/app/grpc"
	"urlSh/internal/config"
	"urlSh/internal/services"
	"urlSh/internal/storage/mongodb"
	"urlSh/internal/storage/redis"
)

type App struct {
	GRPCServer *grpcapp.App
}

func New(log *slog.Logger, cfg *config.Config) *App {
	storage, err := mongodb.New(cfg.Storage.Path, cfg.Storage.Database, cfg.Storage.Collection)
	if err != nil {
		panic(err)
	}
	cache, err := redis.New(cfg.CachePath)
	urlService := services.New(log, storage, cache, cfg.Ttl)

	grpcApp := grpcapp.New(log, cfg, urlService)

	return &App{
		GRPCServer: grpcApp,
	}
}
