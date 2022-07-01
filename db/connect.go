package db

import (
	"database/sql"
	"log"

	"github.com/ygo-skc/skc-suggestion-engine/env"
)

var (
	skcDBConn *sql.DB
)

// Connect to SKC database.
func EstablishSKCDBConn() {
	dataSourceName := env.EnvMap["SKC_DB_USER"] + ":" + env.EnvMap["SKC_DB_PWD"] + "@tcp(" + env.EnvMap["SKC_DB_URI"] + ")/" + env.EnvMap["SKC_DB_NAME"]

	var err error
	if skcDBConn, err = sql.Open("mysql", dataSourceName); err != nil { // root:PWD@tcp(skc-api-db:3306)/skc_api_db for docker, TODO: use env vars to set this dynamically
		log.Fatalln("Error occurred while trying to establish DB connection: ", err)
	}
}
