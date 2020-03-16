package database

import (
	"fmt"
	"time"

	"github.com/fastbill/go-service-toolkit/v3/observance"

	// import migrate mysql, postgres driver
	_ "github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/jinzhu/gorm"
)

// GormLogrus is a logrus logger that implements the gorm interface for logging.
type GormLogrus struct {
	observance.Logger
}

// Print implements the gorm.LogWriter interface, courtesy of https://gist.github.com/bnadland/2e4287b801a47dcfcc94.
func (g GormLogrus) Print(v ...interface{}) {
	if v[0] == "sql" {
		g.WithFields(observance.Fields{"source": "go-service-toolkit/database"}).Debug(fmt.Sprintf("%v - %v", v[3], v[4]))
	}
	if v[0] == "log" {
		g.WithFields(observance.Fields{"source": "go-service-toolkit/database"}).Debug(fmt.Sprintf("%v", v[2]))
	}
}

// SetupGORM loads the ORM with the given configuration
// The setup includes sending a ping and creating the database if it didn't exist.
// A logger will be activated if logLevel is 'debug'.
func SetupGORM(config Config, logger observance.Logger) (*gorm.DB, error) {
	// We have two connection strings:
	// 1) For connecting to the server (and maybe creating the database)
	// 2) For connecting to the database directly.
	dbName := config.Name
	config.Name = ""
	connectionStringWithoutDatabase := config.ConnectionString()
	config.Name = dbName
	connectionString := config.ConnectionString()

	// Open includes sending a ping.
	db, err := gorm.Open(config.Dialect, connectionStringWithoutDatabase)
	if err != nil {
		return nil, err
	}

	if config.Name != "" {
		// Ensure the DB exists.
		db.Exec(fmt.Sprintf(config.createDatabaseQuery(), config.Name))
		err := db.Close()
		if err != nil {
			return nil, err
		}

		// Connect again with DB name.
		db, err = gorm.Open(config.Dialect, connectionString)
		if err != nil {
			return nil, err
		}
	}

	if logger.Level() == "debug" || logger.Level() == "trace" {
		db.LogMode(true)
		gormLogger := GormLogrus{logger}
		db.SetLogger(gormLogger)
	} else {
		db.LogMode(false)
	}

	// This setting addresses "invalid connection" errors in case of connections being closed by the DB server after the wait_timeout (8h).
	// See https://github.com/go-sql-driver/mysql/issues/657.
	db.DB().SetConnMaxLifetime(3500 * time.Second)

	return db, nil
}
