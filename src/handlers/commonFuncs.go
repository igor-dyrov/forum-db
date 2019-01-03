package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/igor-dyrov/forum-db/src/models"
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

func WriteNotFoundMessage(w http.ResponseWriter, message string) {
	output, err := json.Marshal(models.ResponseMessage{Message: message})
	PanicIfError(err)
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(404)
	w.Write(output)
}
