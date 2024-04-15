package postgres

import (
	"context"
	"errors"
	"fmt"
	"sso/internal/domain/models"
	"sso/internal/storage"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

func (s *Storage) GetAppID(ctx context.Context, appName string) (models.App, error) {
	const op = "storage.postgres.GetApp"
	stmt := `SELECT id, name FROM apps WHERE name = $1`

	var app models.App
	err := s.db.QueryRow(ctx, stmt, appName).Scan(&app.ID, &app.Name)
	if err != nil {
		if IsNotFoundError(err) {
			return models.App{}, fmt.Errorf("%s: %w", op, storage.ErrAppNotFound)
		}
		return models.App{}, fmt.Errorf("%s: %w", op, err)
	}
	return app, nil
}

func (s *Storage) GetApp(ctx context.Context, appName string) (models.App, error) {
	const op = "storage.postgres.GetApp"
	stmt := "SELECT id, name, secret FROM apps WHERE name = $1"
	var app models.App
	err := s.db.QueryRow(ctx, stmt, appName).Scan(&app.ID, &app.Name, &app.Secret)
	if err != nil {
		if IsNotFoundError(err) {
			return models.App{}, fmt.Errorf("%s: %w", op, storage.ErrAppNotFound)
		}
		return models.App{}, fmt.Errorf("%s: %w", op, err)
	}
	return app, nil
}

func (s *Storage) SetApp(ctx context.Context, appName string, appSecret string) (int, error) {
	const op = "storage.postgres.SetApp"

	stmt := `INSERT INTO apps (name, secret) VALUES($1, $2) RETURNING id`
	var appID int
	err := s.db.QueryRow(ctx, stmt, appName, appSecret).Scan(&appID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return 0, fmt.Errorf("%s: %w", op, storage.ErrAppExists)
		}
		return 0, fmt.Errorf("%s: %w", op, err)
	}
	return appID, nil
}

func (s *Storage) UpdApp(ctx context.Context, appName, newAppName, newAppSecret string) error {
	const op = "storage.postgres.UdpApp"

	stmt := `SELECT id FROM apps WHERE name = $1`
	var appID int
	err := s.db.QueryRow(ctx, stmt, appName).Scan(&appID)
	if err != nil {
		if IsNotFoundError(err) {
			return fmt.Errorf("%s: %w", op, storage.ErrAppNotFound)
		}
		return fmt.Errorf("%s: %w", op, err)
	}

	stmt = `UPDATE apps SET name = $1, secret = $2 WHERE id = $3`
	_, err = s.db.Exec(ctx, stmt, newAppName, newAppSecret, appID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return fmt.Errorf("%s: %w", op, storage.ErrAppExists)
		}
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (s *Storage) DelApp(ctx context.Context, appName string) error {
	const op = "storage.postgres.DelApp"

	stmt := `SELECT id FROM apps WHERE name = $1`
	var appID int
	err := s.db.QueryRow(ctx, stmt, appName).Scan(&appID)
	if err != nil {
		if IsNotFoundError(err) {
			return fmt.Errorf("%s: %w", op, storage.ErrAppNotFound)
		}
		return fmt.Errorf("%s: %w", op, err)
	}

	stmt = `DELETE FROM creators WHERE app_id = $1`
	_, err = s.db.Exec(ctx, stmt, appID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	stmt = `DELETE FROM admins WHERE app_id = $1`
	_, err = s.db.Exec(ctx, stmt, appID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	stmt = `DELETE FROM apps WHERE name = $1`
	_, err = s.db.Exec(ctx, stmt, appName)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
