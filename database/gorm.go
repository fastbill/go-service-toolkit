package database

import (
	"fmt"
	"time"

	"github.com/fastbill/go-service-toolkit/v4/observance"

	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// SetupGORM loads the ORM with the given configuration
// The setup includes sending a ping and creating the database if it didn't exist.
// A logger will be activated if logLevel is 'debug'.
func SetupGORM(config Config, logger observance.Logger) (*gorm.DB, error) {
	dbName := config.Name

	// First we connect without the database name so we can create the database if it does not exist.
	config.Name = ""
	driverWithoutDatabaseSet := config.Driver()

	gormConfig := &gorm.Config{
		Logger: createLogger(logger),
	}

	// Open includes sending a ping.
	db, err := gorm.Open(driverWithoutDatabaseSet, gormConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to open DB connection: %w", err)
	}

	if dbName != "" {
		config.Name = dbName
		// Ensure the DB exists.
		db.Exec(fmt.Sprintf(config.createDatabaseQuery(), config.Name))
		err = Close(db)
		if err != nil {
			return nil, fmt.Errorf("failed to close DB connection: %w", err)
		}

		// Connect again with DB name.
		driver := config.Driver()
		db, err = gorm.Open(driver, gormConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to open DB connection: %w", err)
		}
	}

	dbConn, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve DB connection: %w", err)
	}

	// This setting addresses "invalid connection" errors in case of connections being closed by the DB server after the wait_timeout (8h).
	// See https://github.com/go-sql-driver/mysql/issues/657.
	dbConn.SetConnMaxLifetime(time.Hour)

	return db, nil
}

// Close closes the database connection(s) used by GORM.
func Close(db *gorm.DB) error {
	dbConn, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to retrieve DB connection: %w", err)
	}

	return dbConn.Close()
}

// GormWriter implements the Writer interface for setting up the GORM logger.
type GormWriter struct {
	observance.Logger
}

// Printf writes a log entry.
func (g GormWriter) Printf(msg string, data ...interface{}) {
	g.Logger.Debug(fmt.Sprintf(msg, data...))
}

func createLogger(logger observance.Logger) gormlogger.Interface {
	var logLevel gormlogger.LogLevel
	if logger.Level() == "debug" || logger.Level() == "trace" {
		logLevel = gormlogger.Info
	} else {
		logLevel = gormlogger.Silent
	}

	newLogger := gormlogger.New(
		GormWriter{Logger: logger},
		gormlogger.Config{
			LogLevel: logLevel,
			Colorful: false,
		},
	)

	return newLogger
}
