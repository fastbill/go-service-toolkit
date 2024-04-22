package envloader

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadingDefaults(t *testing.T) {
	defer cleanup(t)
	err := LoadEnvs("testdata")
	if assert.NoError(t, err) {
		assert.Equal(t, "defaultValue1", os.Getenv("ENVLOADER_TESTKEY1"))
		assert.Equal(t, "defaultValue2", os.Getenv("ENVLOADER_TESTKEY2"))
	}
}

func TestOverwritingDefaultsWithCustomsSuccess(t *testing.T) {
	defer cleanup(t)
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
}

func TestNotOverwritingExistingEnvs(t *testing.T) {
	defer cleanup(t)
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
}

func TestOverwritingDefaultsWithCustomsFail(t *testing.T) {
	defer cleanup(t)
	DefaultEnvFile = "production_missing.env"
	StageEnv = "ENVLOADER_APP_ENV"
	err := os.Setenv("ENVLOADER_APP_ENV", "teststage_fail")
	assert.NoError(t, err)
	err = LoadEnvs("testdata")
	if assert.Error(t, err) {
		assert.Equal(t, "environment variables missing: [ENVLOADER_TESTKEY3]", err.Error())
	}
}

func TestOverwriteEmptyDefaultWithEmptyValue(t *testing.T) {
	defer cleanup(t)
	DefaultEnvFile = "production_missing.env"
	StageEnv = "ENVLOADER_APP_ENV"
	err := os.Setenv("ENVLOADER_APP_ENV", "empty_overwrite")
	assert.NoError(t, err)
	err = LoadEnvs("testdata")
	assert.NoError(t, err)
	assert.Equal(t, "", os.Getenv("ENVLOADER_TESTKEY3"))
}

func TestMissingEnvsAreDetected(t *testing.T) {
	defer cleanup(t)
	DefaultEnvFile = "production_missing.env"
	StageEnv = "ENVLOADER_APP_ENV"
	err := os.Setenv("ENVLOADER_APP_ENV", "missing")
	assert.NoError(t, err)
	err = LoadEnvs("testdata")
	assert.Error(t, err)
	assert.Equal(t, "environment variables missing: [ENVLOADER_TESTKEY3]", err.Error())
}

func cleanup(t *testing.T) {
	for _, line := range os.Environ() {
		if strings.HasPrefix(line, "ENVLOADER") {
			pair := strings.Split(line, "=")
			assert.NoError(t, os.Unsetenv(pair[0]))
		}
	}
}
