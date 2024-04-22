package database

import (
	"errors"
	"fmt"
	"io/ioutil"

	"gorm.io/gorm"
)

// MustApplyDatabaseSeed applies all SQL queries from the given file to the currently active database.
// If any database table besides "schema_migrations" already contains data, the seed file will not be applied.
// The third argument is optional and can be used to exclude tables from checking whether data is already seeded or not.
func MustApplyDatabaseSeed(file string, db *gorm.DB, excludedTables ...[]string) {
	if len(excludedTables) > 1 {
		panic(errors.New("not more than 3 arguments allowed"))
	}

	excludedTablesInSQL := []string{"schema_migrations"}
	if len(excludedTables) == 1 {
		excludedTablesInSQL = append(excludedTablesInSQL, excludedTables[0]...)
	}

	applySeedCheckSQL := `
		SELECT
			SUM(TABLE_ROWS) AS rows2
		FROM
			information_schema.TABLES
		WHERE
			TABLE_SCHEMA = ? AND TABLE_NAME NOT IN ?
	`
	result := struct {
		Rows2 uint64
	}{}
	statement := db.Raw(applySeedCheckSQL, db.Migrator().CurrentDatabase(), excludedTablesInSQL)
	if err := statement.Scan(&result).Error; err != nil {
		panic(fmt.Errorf("failed to check whether seed should be applied: %w", err))
	}

	if result.Rows2 > 0 {
		return
	}

	sql, err := ioutil.ReadFile(file) // nolint: gosec
	if err != nil {
		panic(fmt.Errorf("failed to load seed file: %w", err))
	}

	if err := db.Exec(string(sql)).Error; err != nil {
		panic(fmt.Errorf("failed to apply database seed: %w", err))
	}
}
