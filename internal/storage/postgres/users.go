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

func (s *Storage) SaveUser(ctx context.Context, email string, passHash []byte) (uint64, error) {
	const op = "storage.postgres.SaveUser"

	stmt := `INSERT INTO users (email, pass_hash) VALUES($1, $2) RETURNING id`
	var uid uint64
	err := s.db.QueryRow(ctx, stmt, email, passHash).Scan(&uid)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return 0, fmt.Errorf("%s: %w", op, storage.ErrUserExists)
		}
		return 0, fmt.Errorf("%s: %w", op, err)
	}
	return uid, nil
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
