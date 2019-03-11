package main

import (
	"net/http"
	"os"
	"time"

	toolkit "github.com/fastbill/go-service-toolkit"
	"github.com/labstack/echo"
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
		AppName:              "my-test-app",
		LogLevel:             "debug",
		SentryURL:            "", // insert your Sentry DSN
		Version:              "1.0.0",
		MetricsURL:           "", // insert the URL of the Prometheus Pushgateway
		MetricsFlushInterval: 1 * time.Second,
	}
	obs := toolkit.MustNewObs(obsConfig)
	defer obs.PanicRecover()

	// Set up DB connection and run migrations.
	dbConfig := toolkit.DBConfig{
		Dialect:  "mysql",
		Host:     "localhost",
		Port:     "3310",
		User:     "root",
		Password: "my-secret-pw",
		Name:     "test-db",
	}
	db := toolkit.MustSetupDB(dbConfig, obs.Logger)
	defer func() {
		if err := db.Close(); err != nil {
			obs.Logger.WithError(err).Error("failed to close DB connection")
		}
	}()

	toolkit.MustEnsureDBMigrations("migrations", dbConfig)

	// Set up REDIS cache.
	cache := toolkit.MustNewCache("localhost", "6400", "testPrefix")
	defer func() {
		if err := cache.Close(); err != nil {
			obs.Logger.WithError(err).Error("failed to close REDIS connection")
		}
	}()

	// Set up the server.
	e, connectionsClosed := toolkit.MustNewServer(obs.Logger, "")

	// Set up a routes and handlers.
	e.POST("/users", func(c echo.Context) error {
		obs.Logger.Info("incoming request to create new user")

		newUser := &User{}
		err := c.Bind(newUser)
		if err != nil {
			obs.Logger.WithError(err).Warn("invalid request")
			return c.JSON(http.StatusBadRequest, map[string]string{"msg": err.Error()})
		}

		if err := db.Save(newUser).Error; err != nil {
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
