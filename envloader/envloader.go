package envloader

import (
	"fmt"
	"os"
	"path"

	"github.com/joho/godotenv"
)

// StageEnv is the name of the environment variable that determines the stage the app is running on
var StageEnv = "ENV"

// DefaultStageValue sets what stage to use when it is not set
var DefaultStageValue = "prod"

// DefaultEnvFile determines what file contains the the default env values
var DefaultEnvFile = "prod.env"

// LoadEnvs checks if all envs are set and loads envs from the .env files into process envs
// If envs are missing an error is returned that contains the names of all missing envs
func LoadEnvs(folderPath string) error {
	stage := os.Getenv(StageEnv)
	if stage == "" {
		stage = DefaultStageValue
	}

	customConfigPath := path.Join(folderPath, stage+".env")
	defaultConfigPath := path.Join(folderPath, DefaultEnvFile)

	err := checkForMissingEnvs(customConfigPath, defaultConfigPath)
	if err != nil {
		return err
	}

	return godotenv.Load(customConfigPath, defaultConfigPath)
}

// checkForMissingEnvs errors if an env was defined in the default but not set in
// either the default file, custom config file or environment variables.
// Returns nil otherwise.
func checkForMissingEnvs(customConfigPath string, defaultConfigPath string) error {
	envMapCustom, err := godotenv.Read(customConfigPath)
	if err != nil {
		return fmt.Errorf("error reading custom config file: %w", err)
	}
	envMapDefault, err := godotenv.Read(defaultConfigPath)
	if err != nil {
		return fmt.Errorf("error reading default config file: %w", err)
	}

	fmt.Println(envMapCustom)
	fmt.Println(envMapDefault)

	envMapCombined := map[string]string{}
	for key, value := range envMapCustom {
		envMapCombined[key] = value
	}

	for key, value := range envMapDefault {
		if envMapCombined[key] == "" {
			envMapCombined[key] = value
		}
	}

	missingEnvs := []string{}
	for envName, value := range envMapCombined {
		_, keyPresentInCustomEnvs := envMapCustom[envName]
		if value == "" && os.Getenv(envName) == "" && !keyPresentInCustomEnvs {
			missingEnvs = append(missingEnvs, envName)
		}
	}
	if len(missingEnvs) > 0 {
		return fmt.Errorf("environment variables missing: %v", missingEnvs)
	}

	return nil
}
