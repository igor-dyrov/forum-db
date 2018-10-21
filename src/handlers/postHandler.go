package handlers

import (
	"net/http"
	"io/ioutil"
	"encoding/json"
	"../models"
	"../common"
	"../getters"
	"github.com/gorilla/mux"
	"strconv"
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

	id, err := strconv.Atoi(forum)
	var threadById string
	if  err == nil {
		if !getters.ThreadExists(id) {
			var msg models.ResponseMessage
			msg.Message = `Can't find post thread by id: ` + forum
			output, err := json.Marshal(msg)
			if err != nil {
				http.Error(w, err.Error(), 500)
				return
			}
			w.Header().Set("content-type", "application/json")
			w.WriteHeader(404)
			w.Write(output)
			return
		}
		threadById = getters.GetSlugById(id)
	}

	var Thread int

	if threadById != "" {
		forum = threadById
		Thread = id
	} else {
		forum = getters.GetThreadSlug(forum)
		Thread = getters.GetThreadId(forum)
		if Thread == -1 {
			var msg models.ResponseMessage
			msg.Message = `Can't find post thread by slug: ` + forum
			output, err := json.Marshal(msg)
			if err != nil {
				http.Error(w, err.Error(), 500)
				return
			}
			w.Header().Set("content-type", "application/json")
			w.WriteHeader(404)
			w.Write(output)
			return
		}
	}

	err = nil

	for i := range posts {
		posts[i].Forum = forum
		posts[i].Thread = Thread
		if !getters.UserExists(posts[i].Author) {
			var message models.ResponseMessage
			message.Message = "Can't find thread author by nickname: " + posts[i].Author
			output, err := json.Marshal(message)
			if err != nil {
				http.Error(w, err.Error(), 500)
				return
			}
			w.Header().Set("content-type", "application/json")
			w.WriteHeader(404)
			w.Write(output)
			return
		}
		db.QueryRow(`INSERT INTO posts (author, created, forum, message, thread) VALUES ($1, $2, $3, $4, $5) RETURNING id`,
			posts[i].Author, posts[i].Created, posts[i].Forum, posts[i].Message, posts[i].Thread).Scan(&posts[i].Id)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
	}


	if len(posts) == 0 {
		var post models.Post
		post.Thread = Thread
		post.Forum = forum
		db.QueryRow(`INSERT INTO posts (forum, thread) VALUES($1, $2) RETURNING id`, post.Forum, post.Thread).Scan(&post.Id)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		_, err = db.Exec("UPDATE forums SET posts = posts + 1 WHERE slug = $1", forum)
	} else {
		_, err = db.Exec("UPDATE forums SET posts = posts + $1 WHERE slug = $2", len(posts), forum)
	}

	if err != nil {
		http.Error(w, err.Error(), 500)
		return
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
