package handlertest

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	toolkit "github.com/fastbill/go-service-toolkit/v4"
	"github.com/fastbill/go-service-toolkit/v4/observance"
)

// Suite holds the general properties to run handler tests.
type Suite struct {
	DefaultMiddleware []echo.MiddlewareFunc
	DefaultHeaders    map[string]string
}

// Params contains all settings that should be used when calling the handler.
type Params struct {
	Route             string
	Method            string
	Body              interface{}
	Headers           map[string]string
	PathParams        []PathParam
	Query             map[string]string
	Middleware        []echo.MiddlewareFunc
	SleepBeforeAssert time.Duration
}

// MockAsserter is the interface that needs to be fulfilled by the mocks that are passed in.
type MockAsserter interface {
	AssertExpectations(t mock.TestingT) bool
}

// PathParam defines the name and value for a parameter in the URL.
type PathParam struct {
	Name  string
	Value string
}

// CallHandler calls the given handler with the provided settings.
// Afterwards it calls "AssertExpectations" of the provided mocks.
func (s *Suite) CallHandler(t *testing.T, handlerFunc echo.HandlerFunc, params *Params, mocks []MockAsserter) (*httptest.ResponseRecorder, error) {
	t.Helper()
	if handlerFunc == nil {
		return nil, errors.New("error in test setup, handler missing")
	}

	if params == nil {
		params = &Params{}
	}
	route, method := defineRouteAndMethod(params, t)

	bodyAsReader, err := convertToReader(params.Body)
	assert.NoError(t, err, "error in test setup in CallHandler")

	req := httptest.NewRequest(method, route, bodyAsReader)
	addHeaders(req, s.DefaultHeaders)
	addHeaders(req, params.Headers)
	rec := httptest.NewRecorder()

	obs := &observance.Obs{Logger: observance.NewTestLogger()}
	e, _ := toolkit.MustNewServer(obs, "")
	ctx := e.NewContext(req, rec)

	addPathParams(ctx, params.PathParams)

	allMiddleware := append(s.DefaultMiddleware, params.Middleware...)
	handlerFuncWithMiddleware := applyMiddleware(handlerFunc, allMiddleware)

	defer func() {
		if params.SleepBeforeAssert != 0 {
			time.Sleep(params.SleepBeforeAssert)
		}

		for _, m := range mocks {
			m.AssertExpectations(t)
		}
	}()
	return rec, handlerFuncWithMiddleware(ctx)
}

func convertToReader(body interface{}) (io.Reader, error) {
	if body == nil {
		return nil, nil
	}

	switch v := body.(type) {
	case io.Reader:
		return v, nil
	case string:
		return strings.NewReader(v), nil
	case []byte:
		return bytes.NewReader(v), nil
	default:
		return nil, fmt.Errorf("unknown type %T of body", v)
	}
}

func defineRouteAndMethod(params *Params, t *testing.T) (string, string) {
	method := params.Method
	if method == "" {
		method = "GET"
	}

	route := params.Route
	if route == "" {
		route = "/"
	}

	u, err := url.Parse(route)
	require.NoError(t, err, "error parsing route")
	values, err := url.ParseQuery(u.RawQuery)
	require.NoError(t, err, "error parsing query params in route")
	if params.Query != nil {
		for name, value := range params.Query {
			values.Set(name, value)
		}
		u.RawQuery = values.Encode()
		route = u.String()
	}
	return route, method
}

func addHeaders(req *http.Request, headers map[string]string) {
	for name, value := range headers {
		req.Header.Set(name, value)
	}
}

func addPathParams(ctx echo.Context, pathParams []PathParam) {
	if pathParams == nil {
		return
	}

	names := []string{}
	values := []string{}
	for _, pathParam := range pathParams {
		names = append(names, pathParam.Name)
		values = append(values, pathParam.Value)
	}

	ctx.SetParamNames(names...)
	ctx.SetParamValues(values...)
}

func applyMiddleware(handlerFunc echo.HandlerFunc, middlewares []echo.MiddlewareFunc) echo.HandlerFunc {
	if len(middlewares) == 0 {
		return handlerFunc
	}

	resultFn := handlerFunc
	for i := len(middlewares) - 1; i >= 0; i-- {
		resultFn = middlewares[i](resultFn)
	}
	return resultFn
}
