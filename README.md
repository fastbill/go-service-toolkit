# Service Toolkit [![Build Status](https://travis-ci.com/fastbill/go-service-toolkit.svg?branch=master)](https://travis-ci.com/fastbill/go-service-toolkit) [![Go Report Card](https://goreportcard.com/badge/github.com/fastbill/go-service-toolkit)](https://goreportcard.com/report/github.com/fastbill/go-service-toolkit) [![GoDoc](https://godoc.org/github.com/fastbill/go-service-toolkit?status.svg)](https://godoc.org/github.com/fastbill/go-service-toolkit)

The service toolkit bundles configuration manangement, setting up logging, ORM, REDIS cache and configuring the web framework. It uses opinionated (default) settings to reduce the amount of boilerplate code needed for these tasks. With the toolkit a new Go mircoservice can be set up very quickly.

See [main.go in the example folder](https://github.com/fastbill/go-service-toolkit/blob/master/example/main.go) for a full, working example.

# Configuration and Environment Variables
Following the [12-Factor App Guideline](https://12factor.net/config) our service retrieves its configuration from the environment variables. To avoid having to pass a lot of variables that change rarely or never we keep most values in `.env` files that are then loaded
into environment variables by the envloader package. Values from these files serve as default and are overwritten by values from the environment.

You need to tell the envloader in which folder to look for the `.env` files. By default it will only load the `prod.env` file from that folder. If the environment variable `ENV` is set to e.g. `dev` the the loader will load `dev.env` first and only load additional values not set in there from `prod.env`.

## Usage
```go
import (
	toolkit "github.com/fastbill/go-service-toolkit/v2"
)


func main() {
    // ATTENTION: This needs to be called before any other function from the toolkit is used to ensure the environment variables are correct.
    tookit.MustLoadEnvs("config")
}
```

# Observability - Logging and Metrics
We bundle logging and capturing custom metrics in one `Obs` struct (short for observance). In the future tools for tracing might also be added. Due to the bundling only one struct needs to be passed around in the application and not 2 or 3. Additionally the observance struct provides a method to create request specific observance instances that automatically add url, path and request id to every log message created with that instance. It also adds the account id if it was found in the request header. The headers that are checked for this can be configured via `observance.RequestIDHeader` and `observance.AccountIDHeader`.

We use [Logrus](https://github.com/sirupsen/logrus) as logger under the hood but it is wrapped with a custom interface so we do not depend directly on the interface provided by Logrus. Logs will be written to StdOut in JSON format. If you pass a Sentry URL and version all log entries with level error or higher will be pushed to Sentry. This is done via hooks in Logrus.

The `Obs` struct has a `PanicRecover` method that can be used as deferred function in your setup. It will log the stack trace in case a panic happens.

## Usage
```go
import (
	"time"

	"github.com/fastbill/go-service-toolkit/v2"
)

func main() {
	obsConfig := toolkit.ObsConfig{
		AppName:              "my-test-app",
		LogLevel:             "debug", // required
		SentryURL:            "https://xyz:xyz@sentry.com/123",
		Version:              "1.0.0",
		MetricsURL:           "http://example.com",
		MetricsFlushInterval: 1 * time.Second,
	}

	obs := toolkit.MustNewObs(obsConfig)
	defer obs.PanicRecover()
}
```

The request specific observance is best created in a middleware function that creates a custom context like this:
```go
func SetupCustomContext(obs *observance.Observance) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			return next(&context{
				c,
				obs.CopyWithRequest(c.Request()),
			})
		}
	}
}
```

For testing there is a test logger provided. See the example [here](https://godoc.org/github.com/fastbill/go-service-toolkit/observance#example-NewTestLogger) to find out how to use it.

TODO: Add metrics usage example

# Database
The toolkit allows to set up the database (MySQL or PostgreSQL). The `MustSetupDB` includes the following things:
* Create a database connection
* Check it works via sending a ping
* Create the database with the given name in case it did not exist yet
* Set up the ORM: [GORM](http://gorm.io/)

Additionally `MustEnsureDBMigrations` runs all migrations from the given folder that are missing so far. For that the package [migrate](https://github.com/golang-migrate/migrate) is used.

## Usage
```go
import (
    "github.com/fastbill/go-service-toolkit/v2"
)

func main() {
	dbConfig := toolkit.DBConfig{
		Dialect:  "mysql",
		Host:     "localhost",
		Port:     "3306",
		User:     "root",
		Password: "***",
		Name:     "test-db",
	}
	db := toolkit.MustSetupDB(dbConfig, obs.Logger)
	defer func() {
		if err := db.Close(); err != nil {
			// log the error
		}
	}()

	toolkit.MustEnsureDBMigrations("migrations", dbConfig)
}
```

# Redis Cache
The function `MustNewCache` sets up a new REDIS client. A prefix can be provided that will be added to all keys. The client includes methods to work with JSON data.

## Usage
```go
import (
    "github.com/fastbill/go-service-toolkit/v2"
)

func main() {
	cache := toolkit.MustNewCache("localhost", "6400", "testPrefix")
	defer func() {
		if err := cache.Close(); err != nil {
			// log the error
		}
	}()
}
```

# Server
The server package sets up an [Echo](https://echo.labstack.com/) server that includes graceful shutdown, timeouts, CORS, an error handler that can handle [HTTPErrors](https://github.com/fastbill/httperrors) etc.

The default configuration includes a custom `Bind` method for the context object that performs the default Echo `Bind` but also validates the input via [github.com/go-playground/validator](https://github.com/go-playground/validator).

For the graceful shutdown of the server to work correctly, you need to wait for the channel to be closed at the end of your program as shown below.

## Usage
```go
import (
    "github.com/fastbill/go-service-toolkit/server"
)

func main() {
    echoServer, connectionsClosed := server.MustSetup(logger)
	
	// Set up routes etc.

	err := echoServer.Start(":8080")
	if err != nil {
		obs.Logger.Warn(err)
	}
    <-connectionsClosed
}
```

# Handlertest
This package helps with testing the echo handlers by providing a `CallHandler` method. It allows to specifiy default headers and middleware that should be applied for all handler tests.
Addionally you can define the following parameters that should be applied when the handler function is called.
* Route
* Method
* Body (can be `string`, `[]byte` or `io.Reader`)
* Query parameters (they will be added to the query parameters in the route and overwrite the value of a particular parameter if it already exists)
* Headers (they will overwrite the values that were set in the default headers)
* Path parameters
* Middleware (will be applied before the default middleware)
* [Testify mocks](https://github.com/stretchr/testify#mock-package) for which should be checked whether their expectations were met after the handler was called

All parameters and default parameters are optional.

As response, the `CallHandler` method returns the error that the echo handler returned and the [response recorder](https://golang.org/pkg/net/http/httptest/#ResponseRecorder).

## Usage
### Minimal Case
```go
import (
	"testing"
    "github.com/fastbill/go-service-toolkit/handlertest"
)

func TestMyHandler(t *testing.T) {
	s := handlertest.Suite{}
	rec, err := s.CallHandler(tNew, myHandler, nil, nil)
	// Do the assertions on the response and the error.
}
```
This will call `myHandler` with the route `\` and the method `GET` without additional headers etc.

### Full Example
```go
import (
	"testing"
    "github.com/fastbill/go-service-toolkit/handlertest"
)

var s = handlertest.Suite{
	DefaultHeaders: map[string]string{
		"Content-Type": "application/json",
	},
	Middleware: []echo.MiddlewareFunc{mwDefault},
}

func mwDefault(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Do something
		return next(c)
	}
}

func TestMyHandler(t *testing.T) {
	mwCustom := func (next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Do something
			return next(c)
		}
	}

	params := handlertest.Params{
		Route: "/?query1=value1",
		Method: "PUT",
		Body: `{"id":123}`,
		Headers: map[string]string{
			"testHeader": "someValue",
		},
		Query: map[string]string{
			"query2": "value2",
		},
		PathParams: []PathParam{
			{Name: "param1", Value: "value1"},
		},
		Middleware: []echo.MiddlewareFunc{mwCustom},
	}

	rec, err := s.CallHandler(t, myHandler, params, []MockAsserter{myMock})
	// Do the assertions on the response and the error.
}
```
