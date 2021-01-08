package database

import (
	"fmt"
	"time"

	"github.com/fastbill/go-service-toolkit/v3/observance"

	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// GormWriter implements the Writer interface for setting up the GORM logger.
type GormWriter struct {
	observance.Logger
}

// Printf writes a log entry.
func (g GormWriter) Printf(msg string, data ...interface{}) {
	fullMsg := msg
	for _, entry := range data {
		msg += fmt.Sprintf(", %+v", entry)
	}
	g.Logger.Info(fullMsg)
}

// SetupGORM loads the ORM with the given configuration
// The setup includes sending a ping and creating the database if it didn't exist.
// A logger will be activated if logLevel is 'debug'.
func SetupGORM(config Config, logger observance.Logger) (*gorm.DB, error) {
	// We have two drivers prepared:
	// 1) For connecting to the server (and maybe creating the database)
	// 2) For connecting to the database directly.
	dbName := config.Name
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

	if config.Name != "" {
		// Ensure the DB exists.
		db.Exec(fmt.Sprintf(config.createDatabaseQuery(), config.Name))
		err = Close(db)
		if err != nil {
			return nil, fmt.Errorf("failed to close DB connection: %w", err)
		}

		// Connect again with DB name.
		config.Name = dbName
		driver := config.Driver()
		db, err = gorm.Open(driver, gormConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to open DB connection: %w", err)
		}
	}

	// TODO find out if we need this
	// if logger.Level() == "debug" || logger.Level() == "trace" {
	// 	db.Debug()
	// }

	dbConn, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve DB connection: %w", err)
	}

	// This setting addresses "invalid connection" errors in case of connections being closed by the DB server after the wait_timeout (8h).
	// See https://github.com/go-sql-driver/mysql/issues/657.
	dbConn.SetConnMaxLifetime(time.Hour)

	return db, nil
}

// Close closes the database instance used by GORM.
func Close(db *gorm.DB) error {
	dbConn, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to retrieve DB connection: %w", err)
	}

	return dbConn.Close()
}

var stringToGormLogLevel = map[string]gormlogger.LogLevel{
	"trace": gormlogger.Info,
	"debug": gormlogger.Info,
	"info":  gormlogger.Info,
	"warn":  gormlogger.Warn,
	"error": gormlogger.Error,
}

func createLogger(logger observance.Logger) gormlogger.Interface {
	logLevel := stringToGormLogLevel[logger.Level()]

	newLogger := gormlogger.New(
		GormWriter{},
		gormlogger.Config{
			SlowThreshold: 500 * time.Millisecond,
			LogLevel:      logLevel,
			Colorful:      false,
		},
	)

	return newLogger
}
