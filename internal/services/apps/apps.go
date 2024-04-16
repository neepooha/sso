package apps

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"github.com/neepooha/sso/internal/domain/models"
	"github.com/neepooha/sso/internal/lib/logger/sl"
	"github.com/neepooha/sso/internal/lib/logging"
	"github.com/neepooha/sso/internal/storage"
)

type Apps struct {
	log               *slog.Logger
	appsSetterDeleter AppsSetterDeleter
	userProvider      UserProvider
	creatorProvider   CreatorProvider
	adminProvider     AdminProvider
}

type AppsSetterDeleter interface {
	GetAppID(ctx context.Context, appName string) (models.App, error)
	GetApp(ctx context.Context, appName string) (models.App, error)
	SetApp(ctx context.Context, appName string, appSecret string) (int, error)
	UpdApp(ctx context.Context, appNameOlnd string, appName string, appSecret string) error
	DelApp(ctx context.Context, appName string) error
}

type UserProvider interface {
	GetUser(ctx context.Context, email string) (models.User, error)
}

type CreatorProvider interface {
	SetCreator(ctx context.Context, uID uint64, appID int) error
	IsCreator(ctx context.Context, uID uint64, appName string) error
}

type AdminProvider interface {
	SetAdmin(ctx context.Context, email string, appName string) error
}

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrNotCreator         = errors.New("user isn't creator")
	ErrAppExists          = errors.New("app exists")
	ErrAppNotFound        = errors.New("app not found")
	ErrUserNotCreator     = errors.New("user not creator")
)

// New returns a new instanse of the Permissions service
func New(log *slog.Logger, appsSetterDeleter AppsSetterDeleter, userProvider UserProvider, creatorProvider CreatorProvider, adminProvider AdminProvider) *Apps {
	return &Apps{
		log:               log,
		appsSetterDeleter: appsSetterDeleter,
		userProvider:      userProvider,
		creatorProvider:   creatorProvider,
		adminProvider:     adminProvider,
	}
}

func (a *Apps) GetAppID(ctx context.Context, appName string) (int, string, error) {
	const op = "apps.GetAppID"
	log := a.log.With(slog.String("op", op))

	log.Info("attempting to get appID")
	app, err := a.appsSetterDeleter.GetAppID(ctx, appName)
	if err != nil {
		if errors.Is(err, storage.ErrAppNotFound) {
			log.Warn("app not found ", sl.Err(err))
			return 0, "", fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
		}
		log.Error("failed to get appID", sl.Err(err))
		return 0, "", fmt.Errorf("%s: %w", op, err)
	}
	log.Info("finded app")
	return app.ID, app.Name, nil
}

func (a *Apps) SetApp(ctx context.Context, email string, appName string, appSecret string) (int, error) {
	const op = "apps.SetApp"
	log := a.log.With(slog.String("op", op))

	user, err := a.userProvider.GetUser(ctx, email)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			log.Warn("user not found", sl.Err(err))
			return 0, fmt.Errorf("%s:%w", op, ErrInvalidCredentials)
		}
		log.Error("failed to find user", sl.Err(err))
		return 0, fmt.Errorf("%s:%w", op, err)
	}

	log.Info("attempting to set app")
	appID, err := a.appsSetterDeleter.SetApp(ctx, appName, appSecret)
	if err != nil {
		if errors.Is(err, storage.ErrAppExists) {
			log.Warn("app exists ", sl.Err(err))
			return 0, fmt.Errorf("%s: %w", op, ErrAppExists)
		}
		log.Error("failed to set app", sl.Err(err))
		return 0, fmt.Errorf("%s: %w", op, err)
	}
	err = a.creatorProvider.SetCreator(ctx, user.ID, appID)
	if err != nil {
		log.Error("failed to set creator", sl.Err(err))
		return 0, fmt.Errorf("%s: %w", op, err)
	}
	err = a.adminProvider.SetAdmin(ctx, email, appName)
	if err != nil {
		log.Error("failed to set creator", sl.Err(err))
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("app added")
	return appID, nil
}

func (a *Apps) UpdApp(ctx context.Context, appName string, NewAppName string, NewAppSecret string) (bool, error) {
	const op = "apps.UpdApp"
	log := a.log.With(slog.String("op", op))

	log.Info("attempting to log in")
	err := logging.Logging(ctx, appName, a.creatorProvider, a.appsSetterDeleter)
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

	log.Info("attempting to update app")
	err = a.appsSetterDeleter.UpdApp(ctx, appName, NewAppName, NewAppSecret)
	if err != nil {
		if errors.Is(err, storage.ErrAppNotFound) {
			if errors.Is(err, logging.ErrAppNotFound) {
				log.Warn("app not found", sl.Err(err))
				return false, fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
			}
			log.Warn("app not found ", sl.Err(err))
			return false, fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
		}
		if errors.Is(err, storage.ErrAppExists) {
			log.Error("name already use", sl.Err(err))
			return false, fmt.Errorf("%s: %w", op, ErrAppExists)
		}
		log.Error("failed to update app", sl.Err(err))
		return false, fmt.Errorf("%s: %w", op, err)
	}
	log.Info("app updated")
	return true, nil
}

func (a *Apps) DelApp(ctx context.Context, appName string) (bool, error) {
	const op = "apps.DelApp"
	log := a.log.With(slog.String("op", op))

	log.Info("attempting to log in")
	err := logging.Logging(ctx, appName, a.creatorProvider, a.appsSetterDeleter)
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

	log.Info("attempting to delete app")
	err = a.appsSetterDeleter.DelApp(ctx, appName)
	if err != nil {
		if errors.Is(err, storage.ErrAppNotFound) {
			log.Warn("app not found ", sl.Err(err))
			return true, fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
		}
		log.Error("failed to update app", sl.Err(err))
		return false, fmt.Errorf("%s: %w", op, err)
	}
	log.Info("app deleted")
	return true, nil
}
