package database

import (
	"errors"
	"fmt"
	"path/filepath"
	"runtime"

	"github.com/golang-migrate/migrate/v4"

	// import migrate mysql, postgres driver
	_ "github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"

	// import reading migrations from files
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// EnsureMigrations checks which migrations from the given folder need to be executed
// It performs all missing migrations
// This implementation does not use an existing db instance since the lock release mechanism in "NewWithDatabaseInstance" is buggy
func EnsureMigrations(folder string, config Config) (returnErr error) {
	databaseURL := config.Dialect + "://" + config.MigrationURL()

	fullPathToMigrations, err := filepath.Abs(folder)
	if err != nil {
		return fmt.Errorf("could not determine absolute path to migrations: %w", err)
	}

	migratePathPrefix := "file:"
	if runtime.GOOS != "windows" {
		migratePathPrefix += "//"
	}
	migrations, err := migrate.New(migratePathPrefix+fullPathToMigrations, databaseURL)
	if err != nil {
		return err
	}

	defer func() {
		sErr, dErr := migrations.Close()
		if (sErr != nil || dErr != nil) && returnErr == nil {
			returnErr = errors.New(sErr.Error() + dErr.Error())
		}
	}()

	err = migrations.Up()
	if err != nil && err != migrate.ErrNoChange {
		return err
	}
	return nil
}
