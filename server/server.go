package server

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/golang-migrate/migrate/v4/v2"
	"github.com/fastbill/go-service-toolkit/v2/observance"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const defaultTimeout = 30 * time.Second

var defaultBinder = echo.DefaultBinder{}

// New creates an echo server instance with the given logger, CORS middleware if CORSOrigins was supplied
// and optionally a timeout setting that is applied for read and write.
func New(logger observance.Logger, CORSOrigins string, timeout ...string) (*echo.Echo, chan struct{}, error) {
	timeoutDuration := defaultTimeout
	if len(timeout) > 0 {
		parsedTimeout, err := time.ParseDuration(timeout[0])
		if err != nil {
			return nil, nil, errors.Wrap(err, "timeout could not be parsed")
		}
		timeoutDuration = parsedTimeout
	}

	echoServer := echo.New()

	// Configure Echo.
	echoServer.HideBanner = true
	echoServer.HidePort = true
	echoServer.Server.ReadTimeout = timeoutDuration
	echoServer.Server.WriteTimeout = timeoutDuration
	echoServer.Server.ReadHeaderTimeout = timeoutDuration
	echoServer.HTTPErrorHandler = HTTPErrorHandler(logger)
	echoServer.Binder = &bindValidator{}
	echoServer.Validator = NewValidator()
	echoServer.Logger = Logger{logger}
	echoServer.DisableHTTP2 = true
	echoServer.Pre(middleware.RemoveTrailingSlash())
	echoServer.Use(middleware.Recover())

	if CORSOrigins != "" {
		origins := strings.Split(CORSOrigins, ",")
		echoServer.Use(middleware.CORSWithConfig(middleware.CORSConfig{
			AllowOrigins: origins,
		}))
	}

	// Set up graceful shutdown.
	connsClosed := make(chan struct{})
	sc := make(chan os.Signal)
	go func() {
		s := <-sc
		logger.WithField("signal", s).Warn("shutting down gracefully")

		c, cancel := context.WithTimeout(context.Background(), 9*time.Second)
		defer cancel()

		err := echoServer.Shutdown(c)
		if err != nil {
			logger.Error(err)
		}
		close(connsClosed)
	}()
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM)

	return echoServer, connsClosed, nil
}

// HTTPErrorHandler retruns an error handler that can be used in echo to overwrite the default Echo error handler.
// It can send responses for echo.HTTPError, htttperrors.HTTPError and standard errors.
// Standard errors also get logged.
func HTTPErrorHandler(logger observance.Logger) func(err error, c echo.Context) {
	return func(err error, c echo.Context) {
		// Log error if it is not an HTTPError or an Echo error.
		needsLogging := !isHTTPOrEchoError(err)
		if needsLogging {
			logger.WithFields(logrus.Fields{
				"url":       c.Request().RequestURI,
				"method":    c.Request().Method,
				"requestId": c.Request().Header.Get("Fastbill-Outer-RequestId"),
				"accountId": c.Request().Header.Get("Fastbill-AccountId"),
			}).Error(err)
		}

		httpError := buildHTTPError(err)

		// Send response.
		if !c.Response().Committed {
			var sendErr error
			if c.Request().Method == "HEAD" {
				sendErr = c.NoContent(httpError.StatusCode)
			} else {
				sendErr = httpError.WriteJSON(c.Response())
			}
			if sendErr != nil {
				c.Logger().Error(err)
			}
		}
	}
}

func buildHTTPError(err error) *httperrors.HTTPError {
	if he, ok := err.(*httperrors.HTTPError); ok {
		return he
	}

	if echoError, ok := err.(*echo.HTTPError); ok {
		return httperrors.New(echoError.Code, echoError.Message)
	}

	return httperrors.New(http.StatusInternalServerError, err)
}

func isHTTPOrEchoError(err error) bool {
	_, isHTTPError := err.(*httperrors.HTTPError)
	_, isEchoError := err.(*echo.HTTPError)
	return isHTTPError || isEchoError
}

type bindValidator struct{}

func (b *bindValidator) Bind(i interface{}, c echo.Context) error {
	err := defaultBinder.Bind(i, c)
	if err != nil {
		return err
	}

	return c.Validate(i)
}
