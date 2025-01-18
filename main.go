package main

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/ygo-skc/skc-suggestion-engine/api"
	"github.com/ygo-skc/skc-suggestion-engine/db"
)

func main() {
	db.EstablishDBConn()
	db.EstablishSKCSuggestionEngineDBConn()
	go api.RunHttpServer()
	select {}
}
