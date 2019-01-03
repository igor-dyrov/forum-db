package handlers

import (
	"encoding/json"
	"net/http"

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

func WriteResponce(w http.ResponseWriter, code int, v interface{}) {
	output, err := json.Marshal(v)
	PanicIfError(err)

	w.Header().Set("content-type", "application/json")
	w.WriteHeader(code)
	w.Write(output)
}
