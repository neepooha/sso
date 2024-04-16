package postgres

import (
	"context"
	"fmt"
	"github.com/neepooha/sso/internal/storage"
)

func (s *Storage) IsCreator(ctx context.Context, userID uint64, appName string) error {
	const op = "storage.postgres.IsCreator"

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

	stmt = `SELECT FROM creators WHERE uid = $1 AND app_id = $2`
	err = s.db.QueryRow(ctx, stmt, userID, appID).Scan()
	if err != nil {
		if IsNotFoundError(err) {
			return fmt.Errorf("%s: %w", op, storage.ErrCreatorNotFound)
		}
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (s *Storage) SetCreator(ctx context.Context, userID uint64, appID int) error {
	const op = "storage.postgres.SetCreator"

	stmt := `INSERT INTO creators (uid, app_id) VALUES($1, $2)`
	_, err := s.db.Exec(ctx, stmt, userID, appID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
