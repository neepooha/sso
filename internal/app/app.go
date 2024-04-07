package app

import (
	"log/slog"
	grpcapp "sso/internal/app/grpc"
	"sso/internal/services/auth"
	"sso/internal/storage/sqlite"
	"time"
)

type App struct {
	GRPCSrv *grpcapp.App
}

func New(log *slog.Logger, grpcHost string, grpcPort string, storagePath string, tokenTTL time.Duration) *App {
	storage, err := sqlite.New(storagePath)
	if err != nil {
		panic(err)
	}
	
	authServer := auth.New(log, storage, storage, storage, tokenTTL)

	grpcApp := grpcapp.New(log, authServer, grpcHost, grpcPort)
	return &App{GRPCSrv: grpcApp}
}
