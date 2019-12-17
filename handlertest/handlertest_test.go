package handlertest

import (
	"errors"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type TestMock struct {
	mock.Mock
}

func (m *TestMock) Do() error {
	args := m.Called()
	return args.Error(0)
}

func TestCallHandler_MinimalVersion(t *testing.T) {
	s := Suite{}
	handlerCalled := false
	handler := func(c echo.Context) error {
		handlerCalled = true
		assert.Equal(t, "/", c.Request().RequestURI)
		assert.Equal(t, "GET", c.Request().Method)
		return nil
	}
	tNew := &testing.T{}

	_, err := s.CallHandler(tNew, handler, nil, nil)
	assert.NoError(t, err)
	assert.True(t, handlerCalled, "handler was not called")
	assert.False(t, tNew.Failed())
}

func TestCallHandler_HeaderCorrect(t *testing.T) {
	s := Suite{
		DefaultHeaders: map[string]string{
			"testName1": "testValue1",
			"testName2": "testValue2",
		},
	}

	handlerCalled := false
	handler := func(c echo.Context) error {
		handlerCalled = true

		assert.Equal(t, "testValue1", c.Request().Header.Get("testName1"))
		assert.Equal(t, "otherValue", c.Request().Header.Get("testName2"))
		assert.Equal(t, "testValue3", c.Request().Header.Get("testName3"))
		return nil
	}

	p := &Params{
		Headers: map[string]string{
			"testName2": "otherValue",
			"testName3": "testValue3",
		},
	}

	tNew := &testing.T{}
	_, err := s.CallHandler(tNew, handler, p, nil)
	assert.NoError(t, err)
	assert.True(t, handlerCalled, "handler was not called")
	assert.False(t, tNew.Failed())
}

func TestCallHandler_MethodAndRouteCorrect(t *testing.T) {
	s := Suite{}

	handlerCalled := false
	handler := func(c echo.Context) error {
		handlerCalled = true

		assert.Equal(t, "/some/route?name1=value1&name2=otherValue&name3=value3", c.Request().RequestURI)
		assert.Equal(t, "PUT", c.Request().Method)
		return nil
	}

	p := &Params{
		Route:  "/some/route?name1=value1&name2=value2",
		Method: "PUT",
		Query: map[string]string{
			"name2": "otherValue",
			"name3": "value3",
		},
	}

	tNew := &testing.T{}
	_, err := s.CallHandler(tNew, handler, p, nil)
	assert.NoError(t, err)
	assert.True(t, handlerCalled, "handler was not called")
	assert.False(t, tNew.Failed())
}

func TestCallHandler_BodyCorrect(t *testing.T) {
	body := `{"test":"testBody"}`
	cases := []interface{}{
		body,
		strings.NewReader(body),
		[]byte(body),
	}

	for i, test := range cases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			s := Suite{
				DefaultHeaders: map[string]string{
					"Content-Type": "application/json",
				},
			}

			p := &Params{
				Body: test,
			}

			handlerCalled := false
			handler := func(c echo.Context) error {
				handlerCalled = true
				body := struct {
					Test string `json:"test"`
				}{}
				err := c.Bind(&body)
				assert.NoError(t, err)
				assert.Equal(t, "testBody", body.Test)
				return nil
			}

			tNew := &testing.T{}
			_, err := s.CallHandler(tNew, handler, p, nil)
			assert.NoError(t, err)
			assert.True(t, handlerCalled, "handler was not called")
			assert.False(t, tNew.Failed())
		})
	}
}

func TestCallHandler_InvalidBody(t *testing.T) {
	s := Suite{
		DefaultHeaders: map[string]string{
			"Content-Type": "application/json",
		},
	}

	handler := func(c echo.Context) error { return nil }

	p := &Params{
		Body: 123,
	}

	tNew := &testing.T{}
	_, err := s.CallHandler(tNew, handler, p, nil)
	assert.NoError(t, err)
	assert.True(t, tNew.Failed())
}

func TestCallHandler_MiddlewaresApplied(t *testing.T) {
	mwDefault := func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Request().Header.Set("middleware1", "default was called")
			return next(c)
		}
	}

	mwCustom := func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Request().Header.Set("middleware2", "custom was called")
			return next(c)
		}
	}

	handlerCalled := false
	handler := func(c echo.Context) error {
		handlerCalled = true
		assert.Equal(t, "default was called", c.Request().Header.Get("middleware1"))
		assert.Equal(t, "custom was called", c.Request().Header.Get("middleware2"))
		return nil
	}

	p := &Params{
		Middleware: []echo.MiddlewareFunc{mwCustom},
	}

	s := Suite{
		DefaultMiddleware: []echo.MiddlewareFunc{mwDefault},
	}

	tNew := &testing.T{}
	_, err := s.CallHandler(tNew, handler, p, nil)
	assert.NoError(t, err)
	assert.True(t, handlerCalled, "handler was not called")
	assert.False(t, tNew.Failed())
}

func TestCallHandler_ErrorPropagation(t *testing.T) {
	s := Suite{}
	handler := func(c echo.Context) error {
		return errors.New("test error")
	}
	tNew := &testing.T{}
	_, err := s.CallHandler(tNew, handler, nil, nil)
	assert.EqualError(t, err, "test error")
}

func TestCallHandler_PathParams(t *testing.T) {
	s := Suite{}
	handlerCalled := false
	handler := func(c echo.Context) error {
		handlerCalled = true
		assert.Equal(t, "value1", c.Param("param1"))
		assert.Equal(t, "value2", c.Param("param2"))
		return nil
	}

	p := &Params{
		PathParams: []PathParam{
			{Name: "param1", Value: "value1"},
			{Name: "param2", Value: "value2"},
		},
	}

	tNew := &testing.T{}
	_, err := s.CallHandler(tNew, handler, p, nil)
	assert.NoError(t, err)
	assert.True(t, handlerCalled, "handler was not called")
	assert.False(t, tNew.Failed())
}

func TestCallHanlder_InvalidHandler(t *testing.T) {
	s := Suite{}
	tNew := &testing.T{}
	_, err := s.CallHandler(tNew, nil, nil, nil)
	assert.EqualError(t, err, "error in test setup, handler missing")
}

func TestCallHandler_MockAssertions(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		s := Suite{}
		someMock := &TestMock{}
		someMock.On("Do").Return(nil)

		handler := func(c echo.Context) error {
			err := someMock.Do()
			assert.NoError(t, err)
			return nil
		}

		tNew := &testing.T{}
		_, err := s.CallHandler(tNew, handler, nil, []MockAsserter{someMock})
		assert.NoError(t, err)
		assert.False(t, tNew.Failed())
	})

	t.Run("failure", func(t *testing.T) {
		s := Suite{}
		someMock := &TestMock{}
		someMock.On("Do").Return(nil)

		handler := func(c echo.Context) error {
			// someMock.Do is not called here so tNew will fail
			return nil
		}

		tNew := &testing.T{}
		_, err := s.CallHandler(tNew, handler, nil, []MockAsserter{someMock})
		assert.NoError(t, err)
		assert.True(t, tNew.Failed())
	})
}

func TestCallHandler_SleepBeforeAssert(t *testing.T) {
	s := Suite{}
	someMock := &TestMock{}
	someMock.On("Do").Return(nil)

	handler := func(c echo.Context) error {
		err := someMock.Do()
		assert.NoError(t, err)
		return nil
	}

	p := &Params{
		SleepBeforeAssert: 100 * time.Millisecond,
	}

	tNew := &testing.T{}
	now := time.Now()
	_, err := s.CallHandler(tNew, handler, p, []MockAsserter{someMock})
	assert.NoError(t, err)
	assert.False(t, tNew.Failed())
	assert.Greater(t, int64(time.Since(now)), int64(p.SleepBeforeAssert))
}
