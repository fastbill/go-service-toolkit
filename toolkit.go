package toolkit

import (
	"github.com/fastbill/go-service-toolkit/cache"
	"github.com/fastbill/go-service-toolkit/database"
	"github.com/fastbill/go-service-toolkit/envloader"
	"github.com/fastbill/go-service-toolkit/observance"
	"github.com/fastbill/go-service-toolkit/server"

	"github.com/jinzhu/gorm"
	"github.com/labstack/echo"
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

// MustNewObs creates a new observalibity instance..
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
func MustNewServer(logger observance.Logger, CORSOrigins string, timeout ...string) (*echo.Echo, chan struct{}) {
	echoServer, connectionsClosed, err := server.New(logger, CORSOrigins, timeout...)
	if err != nil {
		panic(err)
	}
	return echoServer, connectionsClosed
}
