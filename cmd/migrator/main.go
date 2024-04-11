package main

import (
	"errors"
	"flag"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/pgconn"
)

func main() {
	var url, dbname, migrationsPath, migrationsTable string
	flag.StringVar(&url, "url", "", "path to storage")
	flag.StringVar(&dbname, "dbname", "", "name of database")
	flag.StringVar(&migrationsPath, "migrations-path", "", "path to migrations")
	flag.StringVar(&migrationsTable, "migrations-table", "", "path to migrations table")
	flag.Parse()

	if url == "" {
		panic("url is required")
	}
	if migrationsPath == "" {
		panic("migrationsPath-path is required")
	}

	m, err := migrate.New("file://"+migrationsPath, fmt.Sprintf("postgres://%s/%s?sslmode=disable&&?x-migrations-table=%s", url, dbname, migrationsTable))
	if err != nil {
		panic(err)
	}
	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			fmt.Println("no migrations to apply")
			return
		}
		panic(err)
	}
	fmt.Println("migrations applied successfully")
}
