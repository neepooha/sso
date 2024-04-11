package perm

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sso/internal/lib/logger/sl"
	"sso/internal/storage"
)

type Permissions struct {
	log                *slog.Logger
	adminSetterDeleter AdminSetterDeleter
}

type AdminSetterDeleter interface {
	SetAdmin(ctx context.Context, email string) (err error)
	DelAdmin(ctx context.Context, email string) (err error)
}

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserNotFound       = errors.New("user not found")
	ErrAdminExists        = errors.New("user already admin")
	ErrAdminNotFound      = errors.New("admin not found")
)

// New returns a new instanse of the Permissions service
func New(log *slog.Logger, adminSetterDeleter AdminSetterDeleter) *Permissions {
	return &Permissions{
		log:                log,
		adminSetterDeleter: adminSetterDeleter,
	}
}

// Login checks if user with given credentials exists in the system
func (a *Permissions) SetAdmin(ctx context.Context, email string) (bool, error) {
	const op = "perm.SetAdmin"
	log := a.log.With(slog.String("op", op))

	log.Info("attempting to set admin")
	err := a.adminSetterDeleter.SetAdmin(ctx, email)
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

// RegisterNewUser registers new user in the system and returns userID
func (a *Permissions) DelAdmin(ctx context.Context, email string) (bool, error) {
	const op = "perm.DelAdmin"
	log := a.log.With(slog.String("op", op))

	log.Info("attempting to delete admin")

	err := a.adminSetterDeleter.DelAdmin(ctx, email)
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
