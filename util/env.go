package util

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

var EnvMap map[string]string

const (
	ENV_VARIABLE_NAME string = "SKC_SUGGESTION_ENGINE_DOT_ENV_FILE"
)

func SetupEnv() {
	EnvMap = ConfigureEnv()
}

func ConfigureEnv() map[string]string {
	if envFile, isOk := os.LookupEnv(ENV_VARIABLE_NAME); !isOk {
		log.Fatalln("Could not find environment variable", ENV_VARIABLE_NAME, "in path.")
	} else {
		log.Println("Loading env from file", envFile)
		if env, err := godotenv.Read(envFile); err != nil {
			log.Fatalln("Could not load environment file (does it exist?). Terminating program.")
		} else {
			return env
		}
	}

	return nil
}
