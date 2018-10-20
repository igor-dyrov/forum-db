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
	"log"
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
		threadById = getters.GetSlugById(id)
	}

	//log.Println("______________________________________")
	//log.Println(forum)
	//log.Println(threadById)
	//log.Println(getters.GetThreadId(forum))

	for i := range posts {
		if threadById != "" {
			posts[i].Forum = threadById
			posts[i].Thread = id
			log.Println(posts[i].Forum)
		} else {
			posts[i].Forum = getters.GetThreadSlug(forum)
			posts[i].Thread = getters.GetThreadId(forum)
		}
		//var req = `INSERT INTO posts (author, forum, message, thread) VALUES ('`
		//req += posts[i].Author + `','`
		//req += posts[i].Forum + `', '`
		//req += posts[i].Message + `', `
		//req += strconv.Itoa(posts[i].Thread) + `)`
		//log.Println(req)

		db.QueryRow(`INSERT INTO posts (author, created, forum, message, thread) VALUES ($1, $2, $3, $4, $5) RETURNING id`,
			posts[i].Author, posts[i].Created, posts[i].Forum, posts[i].Message, posts[i].Thread).Scan(&posts[i].Id)
	}
	_, err = db.Exec("UPDATE forums SET posts = posts + $1 WHERE slug = $2", len(posts), posts[0].Forum)
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
