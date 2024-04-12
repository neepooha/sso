package app

import (
	"log/slog"
	grpcapp "sso/internal/app/grpc"
	"sso/internal/config"
	"sso/internal/services/auth"
	perm "sso/internal/services/permissions"
	"sso/internal/storage/postgres"
)

type App struct {
	GRPCSrv *grpcapp.App
	Storage *postgres.Storage
}

func New(log *slog.Logger, cfg *config.Config) *App {
	storage, err := postgres.New(cfg)
	if err != nil {
		panic(err)
	}

	authServer := auth.New(log, storage, storage, storage, cfg.TokenTTL)
	permServer := perm.New(log, storage)

	grpcApp := grpcapp.New(log, authServer, permServer, cfg.GRPC.Host, cfg.GRPC.Port)
	return &App{GRPCSrv: grpcApp, Storage: storage}
}
