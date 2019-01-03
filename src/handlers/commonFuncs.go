package handlers

import (
	"github.com/jackc/pgx"
)

func PanicIfError(err error) {
	if err != nil {
		panic(err)
	}
}

func PanicIfErrorAndRollback(err error, tx *pgx.Tx) {
	if err != nil {
		tx.Rollback()
		panic(err)
	}
}
