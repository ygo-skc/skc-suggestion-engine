package main

import (
	"os"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	cUtil "github.com/ygo-skc/skc-go/common/util"
	"github.com/ygo-skc/skc-suggestion-engine/api"
	"github.com/ygo-skc/skc-suggestion-engine/db"
	"github.com/ygo-skc/skc-suggestion-engine/downstream"
	_ "google.golang.org/grpc/encoding/gzip"
)

const (
	ENV_VARIABLE_NAME string = "SKC_SUGGESTION_ENGINE_DOT_ENV_FILE"
)

func init() {
	isCICD := os.Getenv("IS_CICD")
	if isCICD != "true" && !strings.HasSuffix(os.Args[0], ".test") {
		cUtil.ConfigureEnv(ENV_VARIABLE_NAME)
	}
}

func main() {
	downstream.ConnectToYGOService()
	db.EstablishDBConn()
	db.EstablishSKCSuggestionEngineDBConn()
	go api.RunHttpServer()
	select {}
}
