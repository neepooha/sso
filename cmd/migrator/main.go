package main

import (
	"errors"
	"flag"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	var storagePath, migrationsPath, migrationsTable string
	flag.StringVar(&storagePath, "storage-path", "", "path to storage")
	flag.StringVar(&migrationsPath, "migrations-path", "", "path to migrations")
	flag.StringVar(&migrationsTable, "migrations-table", "", "name of migrations table")
	flag.Parse()

	if storagePath == "" {
		panic("storage-path is required")
	}
	if migrationsPath == "" {
		panic("migrationsPath-path is required")
	}

	m, err := migrate.New("file://"+migrationsPath, fmt.Sprintf("sqlite3://%s?x-migratioins-table=%s", storagePath, migrationsTable))
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
