package main

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/ygo-skc/skc-suggestion-engine/api"
	"github.com/ygo-skc/skc-suggestion-engine/db"
	"github.com/ygo-skc/skc-suggestion-engine/util"
)

func main() {
	util.SetupEnv()
	db.EstablishSKCDBConn()
	db.EstablishSKCSuggestionEngineDBConn()
	api.SetupMultiplexer()
}
