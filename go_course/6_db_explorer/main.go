package main

import (
	"database/sql"
	"fmt"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
	"github.com/rs/zerolog/log"
)

var (
	// DSN - data source name
	DSN = "root:pass@tcp(localhost:3306)/explorer?charset=utf8&interpolateParams=true"

	serverPort = "8082"
)

func main() {
	db, err := sql.Open("mysql", DSN)
	if err != nil {
		log.Fatal().Err(err).Msg("open mysql db")
	}
	defer func() { logClose(db, "db connection") }()

	err = db.Ping()
	if err != nil {
		log.Fatal().Err(err).Msg("ping db")
	}

	handler, err := NewDbExplorer(db)
	if err != nil {
		log.Fatal().Err(err).Msg("init handler")
	}

	log.Printf("starting server at :%s", serverPort)
	err = http.ListenAndServe(fmt.Sprintf(":%s", serverPort), handler)
	if err != nil {
		log.Fatal().Err(err).Msg("serve handler")
	}
}
