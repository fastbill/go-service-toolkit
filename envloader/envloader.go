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

	missingEnvs := []string{}
	combinedEnvMap, err := createCombinedEnvMap(customConfigPath, defaultConfigPath)
	if err != nil {
		return err
	}

	for envName, value := range combinedEnvMap {
		if value == "" && os.Getenv(envName) == "" {
			missingEnvs = append(missingEnvs, envName)
		}
	}
	if len(missingEnvs) > 0 {
		return fmt.Errorf("environment variables missing: %v", missingEnvs)
	}
	return godotenv.Load(customConfigPath, defaultConfigPath)
}

func createCombinedEnvMap(customConfigPath string, defaultConfigPath string) (map[string]string, error) {
	envMapCustom, err := godotenv.Read(customConfigPath)
	if err != nil {
		return nil, err
	}
	envMapDefault, err := godotenv.Read(defaultConfigPath)
	if err != nil {
		return nil, err
	}

	envMapCombined := envMapCustom
	for key, value := range envMapDefault {
		if envMapCombined[key] == "" {
			envMapCombined[key] = value
		}
	}
	return envMapCombined, nil
}
