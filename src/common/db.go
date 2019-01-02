package common

import (
	"errors"
	"fmt"
	"log"

	"database/sql"

	_ "github.com/lib/pq"
)

var db *sql.DB

const (
	DB_USER     = "docker"
	DB_PASSWORD = "docker"
	DB_NAME     = "forum"
)

func GetDB() *sql.DB {
	if db == nil {
		dbInfo := fmt.Sprintf("user=%s password=%s dbname=%s host = localhost port = 5432 sslmode=disable",
			DB_USER, DB_PASSWORD, DB_NAME)
		var err error
		db, err = sql.Open("postgres", dbInfo)
		if err != nil {
			log.Println(err)
			return nil
		}
		if db == nil {
			panic(errors.New("Couldn't connect to database: db is nil"))
		}
	}
	return db
}

func CloseDB() {
	db.Close()
}
