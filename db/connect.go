package db

import (
	"database/sql"
	"log"
)

var (
	skcDBConn *sql.DB
)

// Connect to SKC database.
func EstablishSKCDBConn() {
	var err error
	skcDBConn, err = sql.Open("mysql", "root@/skc_api_db") // root:PWD@tcp(skc-api-db:3306)/skc_api_db for docker, TODO: use env vars to set this dynamically

	if err != nil {
		log.Fatalln("Error occurred while trying to establish DB connection: ", err)
	}
}
