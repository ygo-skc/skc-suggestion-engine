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
	skcDBConn, err = sql.Open("mysql", "root@/skc_api_db")

	if err != nil {
		log.Fatalln("Error occurred while trying to establish DB connection: ", err)
	}
}
