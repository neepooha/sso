package postgres

import (
	"context"
	"fmt"
	"github.com/neepooha/sso/internal/storage"
)

func (s *Storage) SetAdmin(ctx context.Context, email string, appName string) error {
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

	stmt = `SELECT id FROM apps WHERE name = $1`
	var appID int
	err = s.db.QueryRow(ctx, stmt, appName).Scan(&appID)
	if err != nil {
		if IsNotFoundError(err) {
			return fmt.Errorf("%s: %w", op, storage.ErrAppNotFound)
		}
		return fmt.Errorf("%s: %w", op, err)
	}

	stmt = `SELECT FROM admins WHERE uid = $1 AND app_id = $2`
	err = s.db.QueryRow(ctx, stmt, uid, appID).Scan()
	if err != nil {
		if !IsNotFoundError(err) {
			return fmt.Errorf("%s: %w", op, storage.ErrAdminExists)
		}
	}

	stmt = `INSERT INTO admins (uid, app_id) VALUES ($1, $2)`
	_, err = s.db.Exec(ctx, stmt, uid, appID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Storage) DelAdmin(ctx context.Context, email string, appName string) error {
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

	stmt = `SELECT id FROM apps WHERE name = $1`
	var appID int
	err = s.db.QueryRow(ctx, stmt, appName).Scan(&appID)
	if err != nil {
		if IsNotFoundError(err) {
			return fmt.Errorf("%s: %w", op, storage.ErrAppNotFound)
		}
		return fmt.Errorf("%s: %w", op, err)
	}

	stmt = `SELECT FROM admins WHERE uid = $1 AND app_id = $2`
	err = s.db.QueryRow(ctx, stmt, uid, appID).Scan()
	if err != nil {
		if IsNotFoundError(err) {
			return fmt.Errorf("%s: %w", op, storage.ErrAdminNotFound)
		}
		return fmt.Errorf("%s1: %w", op, err)
	}

	stmt = `DELETE FROM admins WHERE uid = $1 AND app_id = $2`
	_, err = s.db.Exec(ctx, stmt, uid, appID)
	if err != nil {
		return fmt.Errorf("%s2: %w", op, err)
	}
	return nil
}

func (s *Storage) IsAdmin(ctx context.Context, userID uint64, appName string) error {
	const op = "storage.postgres.IsAdmin"
	stmt := `SELECT id FROM apps WHERE name = $1`
	var appID int
	err := s.db.QueryRow(ctx, stmt, appName).Scan(&appID)
	if err != nil {
		if IsNotFoundError(err) {
			return fmt.Errorf("%s: %w", op, storage.ErrAppNotFound)
		}
		return fmt.Errorf("%s: %w", op, err)
	}

	stmt = `SELECT FROM users WHERE id = $1`
	err = s.db.QueryRow(ctx, stmt, userID).Scan()
	if err != nil {
		if IsNotFoundError(err) {
			return fmt.Errorf("%s: %w", op, storage.ErrUserNotFound)
		}
		return fmt.Errorf("%s: %w", op, err)
	}

	stmt = `SELECT FROM admins WHERE uid = $1 AND app_id = $2`
	err = s.db.QueryRow(ctx, stmt, userID, appID).Scan()
	if err != nil {
		if IsNotFoundError(err) {
			return fmt.Errorf("%s: %w", op, storage.ErrAdminNotFound)
		}
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}
