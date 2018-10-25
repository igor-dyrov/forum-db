package handlers

import (
	"net/http"
	"encoding/json"
	"../common"
	"../models"
)

func GetStatus(w http.ResponseWriter, request *http.Request) {
	status := new (models.Status)
	db := common.GetDB()

	rows := db.QueryRow(`SELECT COUNT(*) FROM users`)
	rows.Scan(&status.User)
	rows = db.QueryRow(`SELECT COUNT(*) FROM forums`)
	rows.Scan(&status.Forum)
	rows = db.QueryRow(`SELECT COUNT(*) FROM threads`)
	rows.Scan(&status.Thread)
	rows = db.QueryRow(`SELECT COUNT(*) FROM posts`)
	rows.Scan(&status.Post)

	output, _ := json.Marshal(status)
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(200)
	w.Write(output)
}

func ClearAll(w http.ResponseWriter, request *http.Request) {
	db := common.GetDB()

	_, err := db.Query("TRUNCATE TABLE users, forums, threads, posts, votes")
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	//_, err := db.Query("TRUNCATE TABLE votes")
	//if err != nil {
	//	http.Error(w, err.Error(), 500)
	//	return
	//}
	//_, err = db.Query("TRUNCATE TABLE posts")
	//if err != nil {
	//	http.Error(w, err.Error(), 500)
	//	return
	//}
	//_, err = db.Query("TRUNCATE TABLE posts")
	//if err != nil {
	//	http.Error(w, err.Error(), 500)
	//	return
	//}
	//_, err = db.Query("TRUNCATE TABLE forums")
	//if err != nil {
	//	http.Error(w, err.Error(), 500)
	//	return
	//}
	//_, err = db.Query("TRUNCATE TABLE users")
	//if err != nil {
	//	http.Error(w, err.Error(), 500)
	//	return
	//}
	w.WriteHeader(200)
}