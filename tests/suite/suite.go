package suite

import (
	"context"
	"net"
	"os"
	"sso/internal/config"
	"testing"

	ssov1 "github.com/neepooha/protos/gen/go/sso"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Suite struct {
	*testing.T
	Cfg        *config.Config
	AuthClient ssov1.AuthClient
}

const (
	grpcHost = "localhost"
)

func New(t *testing.T) (context.Context, *Suite) {
	t.Helper()
	t.Parallel()

	cfg := config.MustLoadByPath(configENV())

	ctx, cancelCtx := context.WithTimeout(context.Background(), cfg.GRPC.Timeout)
	t.Cleanup(func() {
		t.Helper()
		cancelCtx()
	})
	cc, err := grpc.DialContext(context.Background(), grpcAdress(cfg), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("grpc server connection failed: %v", err)
	}
	return ctx, &Suite{
		T:          t,
		Cfg:        cfg,
		AuthClient: ssov1.NewAuthClient(cc),
	}
}

func configENV() string {
	// get configPath from our new env
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		return "../config/local.yaml"
	}
	return configPath
}

func grpcAdress(cfg *config.Config) string {
	return net.JoinHostPort(grpcHost, cfg.GRPC.Port)
}
