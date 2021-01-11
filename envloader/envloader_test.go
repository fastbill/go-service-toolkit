package envloader

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadingDefaults(t *testing.T) {
	err := LoadEnvs("testdata")
	if assert.NoError(t, err) {
		assert.Equal(t, "defaultValue1", os.Getenv("ENVLOADER_TESTKEY1"))
		assert.Equal(t, "defaultValue2", os.Getenv("ENVLOADER_TESTKEY2"))
	}
	assert.NoError(t, cleanup())
}

func TestOverwritingDefaultsWithCustomsSuccess(t *testing.T) {
	DefaultEnvFile = "production_missing.env"
	StageEnv = "ENVLOADER_APP_ENV"
	err := os.Setenv("ENVLOADER_APP_ENV", "teststage_success")
	assert.NoError(t, err)
	err = LoadEnvs("testdata")
	if assert.NoError(t, err) {
		assert.Equal(t, "defaultValue1", os.Getenv("ENVLOADER_TESTKEY1"))
		assert.Equal(t, "customValue2", os.Getenv("ENVLOADER_TESTKEY2"))
		assert.Equal(t, "customValue3", os.Getenv("ENVLOADER_TESTKEY3"))
	}
	assert.NoError(t, cleanup())
}

func TestNotOverwritingExistingEnvs(t *testing.T) {
	err := os.Setenv("ENVLOADER_TESTKEY1", "outerValue1")
	assert.NoError(t, err)
	err = os.Setenv("ENVLOADER_TESTKEY2", "outerValue2")
	assert.NoError(t, err)

	DefaultEnvFile = "production_missing.env"
	StageEnv = "ENVLOADER_APP_ENV"
	err = os.Setenv("ENVLOADER_APP_ENV", "teststage_success")
	assert.NoError(t, err)

	err = LoadEnvs("testdata")
	if assert.NoError(t, err) {
		assert.Equal(t, "outerValue1", os.Getenv("ENVLOADER_TESTKEY1"))
		assert.Equal(t, "outerValue2", os.Getenv("ENVLOADER_TESTKEY2"))
		assert.Equal(t, "customValue3", os.Getenv("ENVLOADER_TESTKEY3"))
	}
	assert.NoError(t, cleanup())
}

func TestOverwritingDefaultsWithCustomsFail(t *testing.T) {
	DefaultEnvFile = "production_missing.env"
	StageEnv = "ENVLOADER_APP_ENV"
	err := os.Setenv("ENVLOADER_APP_ENV", "teststage_fail")
	assert.NoError(t, err)
	err = LoadEnvs("testdata")
	if assert.Error(t, err) {
		assert.Equal(t, "environment variables missing: [ENVLOADER_TESTKEY3]", err.Error())
	}
	assert.NoError(t, cleanup())
}

func TestMissingEnvsAreDetected(t *testing.T) {
	StageEnv = "ENVLOADER_APP_ENV"
	err := os.Setenv("ENVLOADER_APP_ENV", "missing")
	assert.NoError(t, err)
	err = LoadEnvs("testdata")
	assert.Equal(t, "environment variables missing: [ENVLOADER_TESTKEY4]", err.Error())
	assert.NoError(t, cleanup())
}

func cleanup() error {
	for _, line := range os.Environ() {
		if strings.HasPrefix(line, "ENVLOADER") {
			pair := strings.Split(line, "=")
			err := os.Unsetenv(pair[0])
			if err != nil {
				return err
			}
		}
	}
	return nil
}
