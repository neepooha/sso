package postgres

import (
	"context"
	"errors"
	"fmt"
	"sso/internal/config"
	"sso/internal/domain/models"
	"sso/internal/storage"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Storage struct {
	db *pgxpool.Pool
}

func New(cfg *config.Config) (*Storage, error) {
	const op = "storage.postgres.NewStorage"

	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s "+
		"password=	%s dbname=%s sslmode=disable",
		cfg.Storage.Host, cfg.Storage.Port, cfg.Storage.User, cfg.Storage.Password, cfg.Storage.Dbname)

	db, err := pgxpool.New(context.Background(), psqlInfo)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Storage{db: db}, nil
}

func (s *Storage) SaveUser(ctx context.Context, email string, passHash []byte) (uint64, error) {
	const op = "storage.postgres.SaveUser"

	stmt := `INSERT INTO users (email, pass_hash) VALUES($1, $2) RETURNING id`
	lastInsertId := 0
	err := s.db.QueryRow(ctx, stmt, email, passHash).Scan(&lastInsertId)
	if err != nil {
		if IsDuplicatedKeyError(err) {
			return 0, fmt.Errorf("%s: %w", op, storage.ErrUserExists)
		}
		return 0, fmt.Errorf("%s: %w", op, err)
	}
	return uint64(lastInsertId), nil
}

func (s *Storage) GetUser(ctx context.Context, email string) (models.User, error) {
	const op = "storage.postgres.GetUser"
	stmt := `SELECT id, email, pass_hash FROM users WHERE email = $1`

	var user models.User
	err := s.db.QueryRow(context.Background(), stmt, email).Scan(&user.ID, &user.Email, &user.PassHash)
	if err != nil {
		if IsNotFoundError(err) {
			return models.User{}, fmt.Errorf("%s: %w", op, storage.ErrUserNotFound)
		}
		return models.User{}, fmt.Errorf("%s: %w", op, err)
	}

	return user, nil
}

func (s *Storage) IsAdmin(ctx context.Context, userID uint64) (bool, error) {
	const op = "storage.postgres.IsAdmin"
	stmt := `SELECT isAdmin FROM admins WHERE uid = $1`

	var isadmin bool
	err := s.db.QueryRow(ctx, stmt, userID).Scan(&isadmin)
	if err != nil {
		if IsNotFoundError(err) {
			return false, fmt.Errorf("%s: %w", op, storage.ErrUserNotFound)
		}
		return false, fmt.Errorf("%s: %w", op, err)
	}
	return isadmin, nil
}

func (s *Storage) App(ctx context.Context, appID int) (models.App, error) {
	const op = "storage.postgres.App"
	stmt := `SELECT id, name, secret FROM apps WHERE id = $1`

	var app models.App
	err := s.db.QueryRow(ctx, stmt, appID).Scan(&app.ID, &app.Name, &app.Secret)
	if err != nil {
		if IsNotFoundError(err) {
			return models.App{}, fmt.Errorf("%s: %w", op, storage.ErrAppNotFound)
		}
		return models.App{}, fmt.Errorf("%s: %w", op, err)
	}
	return app, nil
}

func (s *Storage) SetAdmin(ctx context.Context, email string) error {
	const op = "storage.postgres.SetAdmin"

	stmt := `SELECT id FROM users WHERE email = $1`
	var uid uint64
	err := s.db.QueryRow(ctx, stmt, email).Scan(&uid)
	if err != nil {
		if IsNotFoundError(err) {
			return fmt.Errorf("%s: %w", op, storage.ErrUserNotFound)
		}
		return fmt.Errorf("%s: %w", op, err)
	}

	stmt = `INSERT INTO admins (id) VALUES ($1)`
	_, err = s.db.Exec(ctx, stmt, uid)
	if err != nil {
		if IsDuplicatedKeyError(err) {
			return fmt.Errorf("%s: %w", op, storage.ErrAdminExists)
		}
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Storage) DelAdmin(ctx context.Context, email string) (error) {
	const op = "storage.postgres.DelAdmin"

	stmt := `SELECT id FROM users WHERE email = $1`
	var uid uint64
	err := s.db.QueryRow(ctx, stmt, email).Scan(&uid)
	if err != nil {
		if IsNotFoundError(err) {
			return fmt.Errorf("%s: %w", op, storage.ErrUserNotFound)
		}
		return fmt.Errorf("%s: %w", op, err)
	}

	stmt = `DELETE FROM admins WHERE id = $1`
	res, err := s.db.Exec(ctx, stmt, uid)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	affect := res.RowsAffected()
	if affect == 0 {
		return fmt.Errorf("%s: %w", op, storage.ErrAdminNotFound)
	}
	return nil
}

func IsDuplicatedKeyError(err error) bool {
	var perr *pgconn.PgError
	if errors.As(err, &perr) {
		return perr.Code == "23505" // error code of duplicate
	}
	return false
}

func IsNotFoundError(err error) bool {
	return err.Error() == "no rows in result set"
}