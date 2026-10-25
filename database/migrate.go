package database

import (
	"database/sql"
	"embed"
	"fmt"

	migrate "github.com/rubenv/sql-migrate"
)

//go:embed migrations/*.sql
var migrationFiles embed.FS

// Migrate menjalankan semua file migrasi yang belum dijalankan
func Migrate(db *sql.DB) error {
	source := &migrate.EmbedFileSystemMigrationSource{
		FileSystem: migrationFiles,
		Root:       "migrations",
	}

	n, err := migrate.Exec(db, "postgres", source, migrate.Up)
	if err != nil {
		return err
	}
	fmt.Printf("migrasi selesai, %d migrasi dijalankan\n", n)
	return nil
}
