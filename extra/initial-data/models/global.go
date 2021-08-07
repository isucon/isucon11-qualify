package models

import (
	"log"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

const (
	defaultImagePath = "./images/default.jpg"
)

var (
	db *sqlx.DB
)

func init() {
	// connect to DB
	var err error
	db, err = sqlx.Open("mysql", "isucon:isucon@tcp(127.0.0.1:3306)/isucondition?parseTime=true&loc=Asia%%2FTokyo")
	if err != nil {
		log.Fatal(err)
	}
}
