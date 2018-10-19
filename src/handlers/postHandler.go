package handlers

import (
	"net/http"
	"io/ioutil"
	"encoding/json"
	"../models"
	"../common"
	"github.com/gorilla/mux"
	"time"
)

func CreatePosts(w http.ResponseWriter, request *http.Request) {
	b, err := ioutil.ReadAll(request.Body)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	defer request.Body.Close()
	var posts []models.Post
	err = json.Unmarshal(b, &posts)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	var forum = mux.Vars(request)["slug_or_id"]
	db := common.GetDB()

	for _,post := range posts {
		post.Forum = forum
		post.Created = time.Now()
		_, err = db.Exec(`INSERT INTO posts (author, created, forum, isEdited, message, parent, thread) VALUES ($1, $2, $3, $4, $5, $6, $7)`,
			post.Author, post.Created, post.Forum, post.IsEdited, post.Message, post.Parent, post.Thread)
	}

	output, err := json.Marshal(posts)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(201)
	w.Write(output)
}
