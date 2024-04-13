package perm

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sso/internal/domain/models"
	"sso/internal/lib/logger/sl"
	"sso/internal/storage"

	"github.com/golang-jwt/jwt"
)

type Permissions struct {
	log                *slog.Logger
	adminSetterDeleter AdminSetterDeleter
	appProvider        AppProvider
}

type AdminSetterDeleter interface {
	SetAdmin(ctx context.Context, email string, appID int) (err error)
	DelAdmin(ctx context.Context, email string, appID int) (err error)
	IsAdmin(ctx context.Context, userID uint64, appID int) (bool, error)
	IsCreator(ctx context.Context, uid uint64, appID int) (bool, error)
}

type AppProvider interface {
	App(ctx context.Context, appID int) (models.App, error)
}

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserNotFound       = errors.New("user not found")
	ErrAdminExists        = errors.New("user already admin")
	ErrAdminNotFound      = errors.New("admin not found")
)

// New returns a new instanse of the Permissions service
func New(log *slog.Logger, adminSetterDeleter AdminSetterDeleter, appProvider AppProvider) *Permissions {
	return &Permissions{
		log:                log,
		adminSetterDeleter: adminSetterDeleter,
		appProvider:        appProvider,
	}
}

func (p *Permissions) SetAdmin(ctx context.Context, email string, appID int) (bool, error) {
	const op = "perm.SetAdmin"
	log := p.log.With(slog.String("op", op))

	log.Info("attempting to set admin")
	err := p.adminSetterDeleter.SetAdmin(ctx, email, appID)
	if err != nil {
		if errors.Is(err, storage.ErrAdminExists) {
			log.Warn("user already admin", sl.Err(err))
			return true, nil
		}
		if errors.Is(err, storage.ErrUserNotFound) {
			log.Warn("user not found", sl.Err(err))
			return false, fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
		}
		log.Error("failed to set admin", sl.Err(err))
		return false, fmt.Errorf("%s: %w", op, err)
	}
	log.Info("user is admin now")
	return true, nil
}

func (p *Permissions) DelAdmin(ctx context.Context, email string, appID int) (bool, error) {
	const op = "perm.DelAdmin"
	log := p.log.With(slog.String("op", op))

	log.Info("attempting to delete admin")
	err := p.adminSetterDeleter.DelAdmin(ctx, email, appID)
	if err != nil {
		if errors.Is(err, storage.ErrAdminNotFound) {
			log.Warn("user already not admin", sl.Err(err))
			return true, nil
		}
		if errors.Is(err, storage.ErrUserNotFound) {
			log.Warn("user not found", sl.Err(err))
			return false, fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
		}
		log.Error("failed to delete admin", sl.Err(err))
		return false, fmt.Errorf("%s: %w", op, err)
	}
	log.Info("user is not admin now")
	return true, nil
}

// IsAdmin checks if user is admin
func (p *Permissions) IsAdmin(ctx context.Context, userID uint64, appID int) (bool, error) {
	const op = "perm.IsAdmin"
	log := p.log.With(slog.String("op", op))

	log.Info("checking if user is admin")
	isAdmin, err := p.adminSetterDeleter.IsAdmin(ctx, userID, appID)
	if err != nil {
		if errors.Is(err, storage.ErrAdminNotFound) {
			log.Warn("user is not admin", sl.Err(err))
			return false, nil
		}
		log.Error("failed to check if user is admin", sl.Err(err))
		return false, fmt.Errorf("%s:%w", op, err)
	}

	log.Info("checked if user is admin", slog.Bool("is_admin", isAdmin))
	return isAdmin, nil
}

func (p *Permissions) IsCreator(ctx context.Context, token string, appID int) (bool, error) {
	const op = "perm.IsCreator"
	log := p.log.With(slog.String("op", op))

	log.Info("checking if user is creator")
	app, err := p.appProvider.App(ctx, appID)
	if err != nil {
		if errors.Is(err, storage.ErrAppNotFound) {
			log.Warn("app not found", sl.Err(err))
			return false, fmt.Errorf("%s:%w", op, ErrInvalidCredentials)
		}
		log.Error("failed to find app", sl.Err(err))
		return false, fmt.Errorf("%s:%w", op, err)
	}

	tokenParsed, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) { return []byte(app.Secret), nil })
	if err != nil {
		log.Warn("failed to parse token", sl.Err(err))
		return false, fmt.Errorf("%s:%w", op, ErrInvalidCredentials)
	}
	claims := tokenParsed.Claims.(jwt.MapClaims)
	log.Info("get user claims", slog.Any("claims", claims))
	uid := uint64(claims["uid"].(float64))

	isCreator, err := p.adminSetterDeleter.IsCreator(ctx, uid, app.ID)
	if err != nil {
		if errors.Is(err, storage.ErrCreatorNotFound) {
			log.Warn("creator is not admin", sl.Err(err))
			return false, nil
		}
		log.Error("failed to check if user is admin", sl.Err(err))
		return false, fmt.Errorf("%s:%w", op, err)
	}

	log.Info("checked if user is creator", slog.Bool("is_creator", isCreator))
	return isCreator, nil
}
