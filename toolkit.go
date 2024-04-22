package toolkit

import (
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"

	"github.com/fastbill/go-service-toolkit/v4/cache"
	"github.com/fastbill/go-service-toolkit/v4/database"
	"github.com/fastbill/go-service-toolkit/v4/envloader"
	"github.com/fastbill/go-service-toolkit/v4/observance"
	"github.com/fastbill/go-service-toolkit/v4/server"
)

// MustLoadEnvs checks and loads environment variables from the given folder.
func MustLoadEnvs(folderPath string) {
	err := envloader.LoadEnvs(folderPath)
	if err != nil {
		panic(err)
	}
}

// ObsConfig aliases observance.Config so it will not be necessary to import the observance package for the setup process.
type ObsConfig = observance.Config

// DBConfig aliases database.Config so it will not be necessary to import the database package for the setup process.
type DBConfig = database.Config

// MustNewObs creates a new observability instance.
// It includes the properties "Logger", a Logrus logger that fulfils the Logger interface
// and "Metrics", a Prometheus Client that fulfils the Measurer interface.
func MustNewObs(config ObsConfig) *observance.Obs {
	obs, err := observance.NewObs(config)
	if err != nil {
		panic(err)
	}
	return obs
}

// MustNewCache creates a new REDIS cache client that fulfils the Cache interface.
func MustNewCache(host string, port string, prefix string) *cache.RedisClient {
	redisCache, err := cache.NewRedis(host, port, prefix)
	if err != nil {
		panic(err)
	}
	return redisCache
}

// MustSetupDB creates a new GORM client.
func MustSetupDB(config DBConfig, logger observance.Logger) *gorm.DB {
	db, err := database.SetupGORM(config, logger)
	if err != nil {
		panic(err)
	}
	return db
}

// MustEnsureDBMigrations checks which migration was the last one that was executed and performs all following migrations.
func MustEnsureDBMigrations(folderPath string, config DBConfig) {
	err := database.EnsureMigrations(folderPath, config)
	if err != nil {
		panic(err)
	}
}

// MustNewServer sets up a new Echo server.
func MustNewServer(obs *observance.Obs, CORSOrigins string, timeout ...string) (*echo.Echo, chan struct{}) {
	echoServer, connectionsClosed, err := server.New(obs, CORSOrigins, timeout...)
	if err != nil {
		panic(err)
	}
	return echoServer, connectionsClosed
}

// CloseDatabase closes the database instance used by GORM.
// db.Close() was removed in GORM v2.
func CloseDatabase(db *gorm.DB) error {
	return database.Close(db)
}
