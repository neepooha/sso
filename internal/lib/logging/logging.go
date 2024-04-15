package logging

import (
	"context"
	"errors"
	"fmt"
	"sso/internal/domain/models"
	"sso/internal/storage"
	"strings"

	"github.com/golang-jwt/jwt"
	"google.golang.org/grpc/metadata"
)

var (
	ErrInvalidCredentials = errors.New("invalid appName or token")
	ErrInternalError      = errors.New("invalid credentials")
	ErrCreatorNotFound    = errors.New("creator not found")
	ErrAppNotFound        = errors.New("app not found")
)

type creatorProvider interface {
	IsCreator(ctx context.Context, userID uint64, appName string) error
}

type appProvider interface {
	GetApp(ctx context.Context, appName string) (models.App, error)
}

func Logging(ctx context.Context, appName string, isCreator creatorProvider, getApp appProvider) error {
	const op = "lib.logging.logging"
	tokenStr, err := ExractToken(ctx)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	app, err := getApp.GetApp(ctx, appName)
	if err != nil {
		if errors.Is(err, storage.ErrAppNotFound) {
			return fmt.Errorf("%s: %w", op, ErrAppNotFound)
		}
		return fmt.Errorf("%s: %w", op, err)
	}
	fmt.Println(app)
	tokenParsed, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) { return []byte(app.Secret), nil })
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	claims := tokenParsed.Claims.(jwt.MapClaims)
	uid := uint64(claims["uid"].(float64))

	err = isCreator.IsCreator(ctx, uid, appName)
	if err != nil {
		if errors.Is(err, storage.ErrAppNotFound) {
			return fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
		}
		if errors.Is(err, storage.ErrUserNotFound) {
			return fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
		}
		if errors.Is(err, storage.ErrCreatorNotFound) {
			return fmt.Errorf("%s: %w", op, ErrCreatorNotFound)
		}
		return err
	}
	return nil
}

func ExractToken(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", errors.New("no headers in request")
	}
	authHeaders, ok := md["authorization"]
	if !ok {
		return "", errors.New("no header in request")
	}
	if len(authHeaders) != 1 {
		return "", errors.New("more than 1 header in request")
	}
	auth := authHeaders[0]
	const prefix = "Bearer "
	if !strings.HasPrefix(auth, prefix) {
		return "", errors.New(`missing "Bearer " prefix in "Authorization" header`)
	}
	if auth[len(prefix):] == "" {
		return "", errors.New(`missing token in "Authorization" header`)
	}
	return auth[len(prefix):], nil
}
