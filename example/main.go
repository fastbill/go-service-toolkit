package main

import (
	"net/http"
	"os"
	"time"

	toolkit "github.com/fastbill/go-service-toolkit/v3"
	"github.com/labstack/echo/v4"
)

// User holds all basic user information.
type User struct {
	ID   uint64 `json:"id"`
	Name string `json:"name" validate:"required"`
	Age  uint64 `json:"age" validate:"gte=18"`
}

func main() {
	// Load environment variables.
	toolkit.MustLoadEnvs("config")

	// Set up observance (logging).
	obsConfig := toolkit.ObsConfig{
		AppName:              os.Getenv("APP_NAME"),
		LogLevel:             os.Getenv("LOG_LEVEL"),
		SentryURL:            os.Getenv("SENTRY_URL"),
		Version:              os.Getenv("APP_VERSION"),
		MetricsURL:           os.Getenv("METRICS_URL"),
		MetricsFlushInterval: 1 * time.Second,
		LoggedHeaders: map[string]string{
			"FastBill-RequestId": "requestId",
		},
	}
	obs := toolkit.MustNewObs(obsConfig)
	defer obs.PanicRecover()

	// Set up DB connection and run migrations.
	dbConfig := toolkit.DBConfig{
		Dialect:  os.Getenv("DB_DIALECT"),
		Host:     os.Getenv("DATABASE_HOST"),
		Port:     os.Getenv("DATABASE_PORT"),
		User:     os.Getenv("DATABASE_USER"),
		Password: os.Getenv("DATABASE_PASSWORD"),
		Name:     os.Getenv("DATABASE_NAME"),
	}
	db := toolkit.MustSetupDB(dbConfig, obs.Logger)
	defer func() {
		if err := toolkit.CloseDatabase(db); err != nil {
			obs.Logger.WithError(err).Error("failed to close DB connection")
		}
	}()

	toolkit.MustEnsureDBMigrations("migrations", dbConfig)

	// Set up REDIS cache.
	cache := toolkit.MustNewCache(os.Getenv("REDIS_HOST"), os.Getenv("REDIS_PORT"), "testPrefix")
	defer func() {
		if err := cache.Close(); err != nil {
			obs.Logger.WithError(err).Error("failed to close REDIS connection")
		}
	}()

	// Set up the server.
	e, connectionsClosed := toolkit.MustNewServer(obs, "")

	// Set up a routes and handlers.
	e.POST("/users", func(c echo.Context) error {
		obs.Logger.Info("incoming request to create new user")

		newUser := &User{}
		err := c.Bind(newUser)
		if err != nil {
			obs.Logger.WithError(err).Warn("invalid request")
			return c.JSON(http.StatusBadRequest, map[string]string{"msg": err.Error()})
		}

		if err = db.Save(newUser).Error; err != nil {
			obs.Logger.WithError(err).Error("failed to save user to DB")
			return c.JSON(http.StatusInternalServerError, map[string]string{"msg": err.Error()})
		}

		// Nonsense cache usage example
		err = cache.SetJSON("latestNewUser", newUser, 0)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"msg": err.Error()})
		}

		return c.NoContent(http.StatusCreated)
	})

	// Start the server.
	port := os.Getenv("PORT")
	obs.Logger.WithField("port", port).Info("server running")
	err := e.Start(":" + port)
	if err != nil {
		obs.Logger.Warn(err)
	}

	<-connectionsClosed // Wait for the graceful shutdown to finish.
}
