package handlers

import (
	"net/http"

	"github.com/igor-dyrov/forum-db/src/common"
	"github.com/igor-dyrov/forum-db/src/models"
)

func GetStatus(w http.ResponseWriter, request *http.Request) {
	status := new(models.Status)

	conn := common.GetConnection()
	defer common.Release(conn)

	rows := conn.QueryRow(`SELECT COUNT(*) FROM users`)
	rows.Scan(&status.User)
	rows = conn.QueryRow(`SELECT COUNT(*) FROM forums`)
	rows.Scan(&status.Forum)
	rows = conn.QueryRow(`SELECT COUNT(*) FROM threads`)
	rows.Scan(&status.Thread)
	rows = conn.QueryRow(`SELECT COUNT(*) FROM posts`)
	rows.Scan(&status.Post)

	WriteResponce(w, 200, status)
}

func ClearAll(w http.ResponseWriter, request *http.Request) {

	conn := common.GetConnection()
	defer common.Release(conn)

	_, err := conn.Exec("TRUNCATE TABLE forum_users, users, forums, threads, posts, votes")
	PanicIfError(err)

	w.WriteHeader(200)
}
