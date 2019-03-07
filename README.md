# Service Toolkit

# Configuration/ Environment variables
Once started, the Go services get their configuration from environment variables. To avoid having to pass a lot of variables that change rarely or never we keep them in `.env` files that are then loaded
into environment variables by this envloader package.

You need to tell the envloader in which folder to look for the `.env` files. By default it will only load the `produ.env` file. If the environment variable `ENV` is set to e.g. `dev` the the loader will load   `dev.env` first and only load additional values not set in there from `prod.env`. If any variable was passed in as environment variable the value will be preserved and not overwritten by the values in the files.

## Usage
```go
import (
	"github.com/fastbill/go-service-toolkit/envloader"
)


func main() {
    // ATTENTION: the envloader needs to be called before any other package from the go-service-setup is used
    envloader.MustSetup("config")
}
```

# Logging and Observance
We use [Logrus](https://github.com/sirupsen/logrus) as logger. The logging package in this repo does the logger setup for you and returns you a logger instance that you can then use throughout the application.
If you pass a Sentry URL and version all log entries with level error or higher will be pushed to Sentry.

The logger can then be wrapped into an `Observance` struct that will hold more things like metrics and tracing later on. The observance struct allows to create request specific observance instances that automatically add url, path and request id to every log message created with that instance. It also adds the account id if it was found in the request header.


## Usage
```go
import (
	"github.com/fastbill/go-service-toolkit/logging"
)

func main() {
    logger := logging.MustSetup("debug", "some-service", "http://someSentryURL", "1.0.0")
    // alternativly without Sentry
    logger := logging.MustSetup("debug", "some-service", "", "")

    // create the general observance object
    obs := &observance.Observance{Logger: logger}
}

 // e.g. inside your middleware, create a request specific observance instance
func Setup(obs *observance.Observance) echo.MiddlewareFunc {
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

# Database
This package sets up the database connection and allows running migrations.

## Usage
```go
import (
    "github.com/fastbill/go-service-toolkit/database"
)

func main() {
    dbConfig := database.Config{
		Dialect:  os.Getenv("DB_DIALECT"),
		Host:     os.Getenv("DATABASE_HOST"),
		Port:     os.Getenv("DATABASE_PORT"),
		User:     os.Getenv("DATABASE_USER"),
		Password: os.Getenv("DATABASE_PASSWORD"),
		Name:     os.Getenv("DATABASE_NAME"),
	}
	db := database.MustSetupGORM(dbConfig, logger, logLevel)
	defer func() { 
        if err := db.Close(); err != nil {
			// log the error
		}
    }()
    database.MustEnsureMigrations("migrations", dbConfig)
}
```

# Redis
This package sets up a connection to REDIS.

## Usage
```go
import (
    "github.com/fastbill/go-service-toolkit/redis"
)

func main() {
    redisClient := redis.MustSetup("127.0.0.1", "6379", "debug")
    defer func() { 
        if err := redisClient.Close(); err != nil {
			// log the error
		}
    }()
}
```

# Server
The server package sets up an [Echo](https://echo.labstack.com/) server that includes graceful shutdown, timeouts, CORS, an error handler that can handle [HTTPErrors](https://github.com/fastbill/httperrors) etc.  
You should also always set up the Echo recover middleware as shown below.  
For the graceful shutdown of the server to work correctly, you need to wait for the channel to return at the end of your program.

## Usage
```go
import (
    "github.com/fastbill/go-service-toolkit/server"
)

func main() {
    echoServer, connectionsClosed := server.MustSetup(logger)
    echoServer.Use(middleware.Recover())
    
    logger.Warn(echoServer.Start(":8080"))
    <- connectionsClosed
}
```


# Validator
The validator package wraps [github.com/go-playground/validator](https://github.com/go-playground/validator) so it becomes compatible with the Echo validation interface.

## Usage 
```go
echoServer.Validator = validator.New()
```
