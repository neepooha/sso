package perm

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sso/internal/domain/models"
	"sso/internal/lib/logger/sl"
	"sso/internal/lib/logging"
	"sso/internal/storage"
)

type Permissions struct {
	log                *slog.Logger
	adminSetterDeleter AdminSetterDeleter
	appProvider        AppProvider
}

type AdminSetterDeleter interface {
	SetAdmin(ctx context.Context, email string, appName string) error
	DelAdmin(ctx context.Context, email string, appName string) error
	IsAdmin(ctx context.Context, userID uint64, appName string) error
	IsCreator(ctx context.Context, userID uint64, appName string) error
}

type AppProvider interface {
	GetApp(ctx context.Context, appName string) (models.App, error)
}

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserNotFound       = errors.New("user not found")
	ErrNotCreator         = errors.New("ErrNotCreator")
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

func (p *Permissions) SetAdmin(ctx context.Context, email string, appName string) (bool, error) {
	const op = "perm.SetAdmin"
	log := p.log.With(slog.String("op", op))

	log.Info("attempting to log in")
	err := logging.Logging(ctx, appName, p.adminSetterDeleter, p.appProvider)
	if err != nil {
		if errors.Is(err, logging.ErrCreatorNotFound) {
			log.Warn("user not creator", sl.Err(err))
			return false, fmt.Errorf("%s: %w", op, ErrNotCreator)
		}
		if errors.Is(err, logging.ErrInvalidCredentials) {
			log.Warn("cant get info of user", sl.Err(err))
			return false, fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
		}
		log.Warn("error logging", sl.Err(err))
		return false, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("attempting to set admin")
	err = p.adminSetterDeleter.SetAdmin(ctx, email, appName)
	if err != nil {
		if errors.Is(err, storage.ErrAdminExists) {
			log.Warn("user already admin", sl.Err(err))
			return true, nil
		}
		if errors.Is(err, storage.ErrUserNotFound) {
			log.Warn("user not found", sl.Err(err))
			return false, fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
		}
		if errors.Is(err, storage.ErrAppNotFound) {
			log.Warn("app not found", sl.Err(err))
			return false, fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
		}
		log.Error("failed to set admin", sl.Err(err))
		return false, fmt.Errorf("%s: %w", op, err)
	}
	log.Info("user is admin now")
	return true, nil
}

func (p *Permissions) DelAdmin(ctx context.Context, email string, appName string) (bool, error) {
	const op = "perm.DelAdmin"
	log := p.log.With(slog.String("op", op))

	log.Info("attempting to log in")
	err := logging.Logging(ctx, appName, p.adminSetterDeleter, p.appProvider)
	if err != nil {
		if errors.Is(err, logging.ErrCreatorNotFound) {
			log.Warn("user not creator", sl.Err(err))
			return false, fmt.Errorf("%s: %w", op, ErrNotCreator)
		}
		if errors.Is(err, logging.ErrInvalidCredentials) {
			log.Warn("cant get info of user", sl.Err(err))
			return false, fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
		}
		log.Warn("error logging", sl.Err(err))
		return false, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("attempting to delete admin")
	err = p.adminSetterDeleter.DelAdmin(ctx, email, appName)
	if err != nil {
		if errors.Is(err, storage.ErrAdminNotFound) {
			log.Warn("user already not admin", sl.Err(err))
			return true, nil
		}
		if errors.Is(err, storage.ErrUserNotFound) {
			log.Warn("user not found", sl.Err(err))
			return false, fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
		}
		if errors.Is(err, storage.ErrAppNotFound) {
			log.Warn("app not found", sl.Err(err))
			return false, fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
		}
		log.Error("failed to delete admin", sl.Err(err))
		return false, fmt.Errorf("%s: %w", op, err)
	}
	log.Info("user is not admin now")
	return true, nil
}

// IsAdmin checks if user is admin
func (p *Permissions) IsAdmin(ctx context.Context, userID uint64, appName string) (bool, error) {
	const op = "perm.IsAdmin"
	log := p.log.With(slog.String("op", op))

	log.Info("checking if user is admin")
	err := p.adminSetterDeleter.IsAdmin(ctx, userID, appName)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			log.Warn("user not found", sl.Err(err))
			return false, fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
		}
		if errors.Is(err, storage.ErrAdminNotFound) {
			log.Warn("user is not admin", sl.Err(err))
			return false, nil
		}
		if errors.Is(err, storage.ErrAppNotFound) {
			log.Warn("app not found", sl.Err(err))
			return false, fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
		}
		log.Error("failed to check if user is admin", sl.Err(err))
		return false, fmt.Errorf("%s:%w", op, err)
	}

	log.Info("checked if user is admin", slog.Bool("is_admin", true))
	return true, nil
}

func (p *Permissions) IsCreator(ctx context.Context, userID uint64, appName string) (bool, error) {
	const op = "perm.IsCreator"
	log := p.log.With(slog.String("op", op))

	log.Info("checking if user is creator")
	_, err := p.appProvider.GetApp(ctx, appName)
	if err != nil {
		if errors.Is(err, storage.ErrAppNotFound) {
			log.Warn("app not found", sl.Err(err))
			return false, fmt.Errorf("%s:%w", op, ErrInvalidCredentials)
		}
		log.Error("failed to find app", sl.Err(err))
		return false, fmt.Errorf("%s:%w", op, err)
	}

	err = p.adminSetterDeleter.IsCreator(ctx, userID, appName)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			log.Warn("user not found", sl.Err(err))
			return false, fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
		}
		if errors.Is(err, storage.ErrCreatorNotFound) {
			log.Warn("user is not creator", sl.Err(err))
			return false, nil
		}
		log.Error("failed to check if user is creator", sl.Err(err))
		return false, fmt.Errorf("%s:%w", op, err)
	}

	log.Info("checked if user is creator", slog.Bool("is_creator", true))
	return true, nil
}
