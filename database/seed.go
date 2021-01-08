package database

import (
	"fmt"
	"io/ioutil"

	"gorm.io/gorm"
)

// MustApplyDatabaseSeed applies all SQL queries from the given file to the currently active database.
// If any database table besides "schema_migrations" already contains data, the seed file will not be applied.
func MustApplyDatabaseSeed(file string, db *gorm.DB) {
	applySeedCheckSQL := `
		SELECT
			SUM(TABLE_ROWS) AS rows
		FROM
			information_schema.TABLES
		WHERE
			TABLE_SCHEMA = ? AND TABLE_NAME NOT IN ('schema_migrations')
	`
	result := struct {
		Rows uint64
	}{}
	if err := db.Raw(applySeedCheckSQL).Scan(&result).Error; err != nil {
		panic(fmt.Errorf("failed to check whether seed should be applied: %w", err))
	}

	if result.Rows > 0 {
		return
	}

	sql, err := ioutil.ReadFile(file)
	if err != nil {
		panic(fmt.Errorf("failed to load seed file: %w", err))
	}

	if err := db.Exec(string(sql)).Error; err != nil {
		panic(fmt.Errorf("failed to apply database seed: %w", err))
	}
}
