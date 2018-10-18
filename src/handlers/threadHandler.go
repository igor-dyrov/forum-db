package handlers

import (
	"net/http"
	"io/ioutil"
	"encoding/json"
	"../models"
	"../common"
	"../getters"
	"github.com/gorilla/mux"
)

func CreateThread(w http.ResponseWriter, request *http.Request) {
	b, err := ioutil.ReadAll(request.Body)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	defer request.Body.Close()
	var thread models.Thread
	err = json.Unmarshal(b, &thread)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	thread.Forum = mux.Vars(request)["slug"]
	db := common.GetDB()
	if thread.Slug != "" {
		_, err = db.Exec("INSERT INTO threads (slug, created, message, title, author, forum, votes) VALUES ($1, $2, $3, $4, $5, $6, $7)",
			thread.Slug, thread.Created, thread.Message, thread.Title, thread.Author, thread.Forum, thread.Votes)
	} else {
		_, err = db.Exec("INSERT INTO threads (created, message, title, author, forum, votes) VALUES ($1, $2, $3, $4, $5, $6)",
			thread.Created, thread.Message, thread.Title, thread.Author, thread.Forum, thread.Votes)
	}
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	thread.Id = getters.GetIdByNickname(thread.Author)
	output, err := json.Marshal(thread)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(201)
	w.Write(output)
}
