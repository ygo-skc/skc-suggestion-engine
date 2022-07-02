package main

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/ygo-skc/skc-suggestion-engine/api"
	"github.com/ygo-skc/skc-suggestion-engine/db"
	"github.com/ygo-skc/skc-suggestion-engine/env"
)

func main() {
	env.LoadEnv()
	db.EstablishSKCDBConn()
	db.EstablishSKCSuggestionEngineDBConn()
	db.Test()
	api.SetupMultiplexer()
}
