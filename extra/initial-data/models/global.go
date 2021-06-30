package models

import (
	"log"
	"net/http"
	"net/url"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

const (
	defaultImagePath = "./images/default.png"
)

var (
	db     *sqlx.DB
	apiUrl *url.URL
)

func init() {
	// connect to DB
	var err error
	db, err = sqlx.Open("mysql", "isucon:isucon@tcp(127.0.0.1:3306)/isucondition?parseTime=true&loc=Local")
	if err != nil {
		log.Fatal(err)
	}

	// connect to API
	apiUrl, err = url.Parse("http://127.0.0.1:3000/")
	if err != nil {
		log.Fatal(err)
	}
	if _, err := http.Get(apiUrl.String()); err != nil {
		log.Fatal(err)
	}
}
