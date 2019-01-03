package common

import (
	"errors"
	"fmt"
	"log"

	"database/sql"

	"github.com/jackc/pgx"
	_ "github.com/lib/pq"
)

var (
	db       *sql.DB
	deadpool *pgx.ConnPool
)

const (
	DB_USER     = "docker"
	DB_PASSWORD = "docker"
	DB_NAME     = "forum"
)

const (
	dbUser     = "docker"
	dbPassword = "docker"
	dbName     = "forum"

	connPoolSize = 42
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

func InitConnectionPool() {
	if deadpool == nil {

		config := pgx.ConnConfig{
			Host:     "localhost",
			User:     dbUser,
			Password: dbPassword,
			Database: dbName,
			Port:     5432,
		}

		conns, err := pgx.NewConnPool(
			pgx.ConnPoolConfig{
				ConnConfig:     config,
				MaxConnections: 42,
			},
		)
		if err != nil {
			fmt.Println("DB connection error: ", err)
			panic(err)
		}

		deadpool = conns
	}
}

func makeConnectionConfig() pgx.ConnConfig {
	return pgx.ConnConfig{
		Host:     "localhost",
		User:     dbUser,
		Password: dbPassword,
		Database: dbName,
		Port:     5432,
	}
}
