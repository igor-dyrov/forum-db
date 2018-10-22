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

	if !getters.UserExists(thread.Author) {
		var message models.ResponseMessage
		message.Message = "Can't find thread author by nickname: " + thread.Author
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

	if !getters.SlugExists(thread.Forum) {
		var message models.ResponseMessage
		message.Message = "Can't find thread forum by slug: " + thread.Forum
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

	var gotThread = getters.GetThreadBySlug(thread.Slug)
	if gotThread.Slug != "" {
		output, err := json.Marshal(gotThread)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		w.Header().Set("content-type", "application/json")
		w.WriteHeader(409)
		w.Write(output)
		return
	}

	if thread.Slug != "" {
		db.QueryRow("INSERT INTO threads (slug, created, message, title, author, forum, votes) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id",
			thread.Slug, thread.Created, thread.Message, thread.Title, thread.Author, thread.Forum, thread.Votes).Scan(&thread.ID)
	} else {
		db.QueryRow("INSERT INTO threads (created, message, title, author, forum, votes) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id",
			thread.Created, thread.Message, thread.Title, thread.Author, thread.Forum, thread.Votes).Scan(&thread.ID)
	}
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	_, err = db.Exec("UPDATE forums SET threads = threads + 1 WHERE slug = $1", thread.Forum)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	thread.Forum = getters.GetSlugCase(thread.Forum)
	output, err := json.Marshal(thread)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(201)
	w.Write(output)
}

func GetThreads(w http.ResponseWriter, request *http.Request) {
	var slug = mux.Vars(request)["slug"]

	limit := request.URL.Query().Get("limit")
	since := request.URL.Query().Get("since")
	desc := request.URL.Query().Get("desc")

	gotThreads := getters.GetThreadsByForum(slug)
	if len(gotThreads) == 0 {
		var msg models.ResponseMessage
		msg.Message = "Can't find forum by slug: " + slug
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

	gotThreads = getters.GetThreads(slug, limit, since, desc)
	output, err := json.Marshal(gotThreads)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(200)
	w.Write(output)
}

func ThreadDetails(w http.ResponseWriter, request *http.Request) {
	var slug_or_id = mux.Vars(request)["slug_or_id"]
	db := common.GetDB()

	var result models.Thread
	result.ID = -1
	id, err := strconv.Atoi(slug_or_id)
	if  err == nil {
		rows, err := db.Query(`SELECT * FROM threads WHERE id = $1`, id)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		for rows.Next() {
			rows.Scan(&result.ID, &result.Slug, &result.Created, &result.Message, &result.Title,
				&result.Author, &result.Forum, &result.Votes)
		}
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		if result.ID == -1 {
			var msg models.ResponseMessage
			msg.Message = "Can`t find thread by id: " + slug_or_id
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
	} else {
		rows, err := db.Query(`SELECT * FROM threads WHERE slug = $1`, slug_or_id)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		for rows.Next() {
			rows.Scan(&result.ID, &result.Slug, &result.Created, &result.Message, &result.Title,
				&result.Author, &result.Forum, &result.Votes)
		}
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		if result.ID == -1 {
			var msg models.ResponseMessage
			msg.Message = "Can`t find thread by slug: " + slug_or_id
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
	output, err := json.Marshal(result)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(200)
	w.Write(output)
}