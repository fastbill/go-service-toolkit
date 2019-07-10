package handlertest

import (
	"testing"

	"github.com/labstack/echo"
	"github.com/stretchr/testify/assert"
)

func TestCallHandler(t *testing.T) {
	t.Run("succes", func(t *testing.T) {
		s := Suite{}
		fn := func(c echo.Context) error {
			return nil
		}
		_, err := s.CallHandler(t, fn, &Params{Body: nil}, nil)
		assert.NoError(t, err)
	})
}
