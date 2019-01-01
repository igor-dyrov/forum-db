package common

import (
	_ "github.com/lib/pq"
	"database/sql"
	"fmt"
	"log"
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
	}
	return db
}
func CloseDB() {
	db.Close()
}