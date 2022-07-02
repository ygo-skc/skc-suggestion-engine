package env

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

var EnvMap map[string]string

const (
	ENV_VARIABLE_NAME string = "SKC_SUGGESTION_ENGINE_DOT_ENV_FILE"
)

func LoadEnv() {
	EnvMap = ConfigureEnv()
}

func ConfigureEnv() map[string]string {
	if envFile, isEnvironmentVariable := os.LookupEnv(ENV_VARIABLE_NAME); !isEnvironmentVariable {
		log.Fatalln("Could not find environment variable", ENV_VARIABLE_NAME, "in path.")
	} else {
		log.Println("Loading env using file", envFile)
		env, err := godotenv.Read(envFile)
		if err != nil {
			log.Fatalln("Could not load environment file (does it exist?). Terminating program.")
		} else {
			return env
		}
	}

	return nil
}
