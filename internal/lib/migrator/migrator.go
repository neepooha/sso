package migrator

import (
	"fmt"
	"github.com/neepooha/sso/internal/config"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/pgconn"
)

func Migrate(cfg *config.Config) error {
	m, err := migrate.New("file://"+cfg.Storage.Migrations_path,
		fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", cfg.Storage.User, cfg.Storage.Password, cfg.Storage.Host, cfg.Storage.Port, cfg.Storage.Dbname))
	if err != nil {
		return err
	}
	if err := m.Up(); err != nil {
		return err
	}
	return nil
}
