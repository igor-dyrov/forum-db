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
	"time"
)

func CreatePosts(w http.ResponseWriter, request *http.Request) {
	var curTime = time.Now()
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
	var slug_or_id = mux.Vars(request)["slug_or_id"]
	db := common.GetDB()

	id, err := strconv.Atoi(slug_or_id) //try to get id
	var forum string
	if  err == nil { //got id
		if !getters.ThreadExists(id) { //check user by id
			var msg models.ResponseMessage
			msg.Message = `Can't find post thread by id: ` + slug_or_id
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
		forum = getters.GetSlugById(id) //get forum by id
	}
	err = nil
	var Thread int
	if forum != "" {
		Thread = id
	} else {
		forum = getters.GetThreadSlug(slug_or_id) //get forum from thread by slug
		Thread = getters.GetThreadId(slug_or_id)
		if Thread == -1 { //can`t find forum by id
			var msg models.ResponseMessage
			msg.Message = `Can't find post thread by slug: ` + slug_or_id
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
		posts[i].Created = curTime
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
		if posts[i].Parent != 0 && !getters.CheckParent(posts[i].Parent, posts[i].Thread) {
			var msg models.ResponseMessage
			msg.Message = `Parent post was created in another thread`
			output, err := json.Marshal(msg)
			if err != nil {
				http.Error(w, err.Error(), 500)
				return
			}
			w.Header().Set("content-type", "application/json")
			w.WriteHeader(409)
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
		post.Created = curTime
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

func GetPost(w http.ResponseWriter, request *http.Request) {
	var id = mux.Vars(request)["id"]
	db := common.GetDB()
	var result models.Post
	result.Id = -1
	rows, err := db.Query(`SELECT * FROM posts WHERE id = $1`, id)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	for rows.Next()  {
		err = rows.Scan(&result.Id, &result.Author, &result.Created, &result.Forum,
			&result.IsEdited, &result.Message, &result.Parent, &result.Thread, &result.Path)
	}
	//if err != nil{
	//	http.Error(w, err.Error(), 500)
	//	return
	//}
	if result.Id == -1 {
		var msg models.ResponseMessage
		msg.Message = `Can't find post post by id: ` + id
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
	var Post models.PostResponse
	Post.Post = result
	output, err := json.Marshal(Post)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(200)
	w.Write(output)
}

func UpdatePost(w http.ResponseWriter, request *http.Request) {
	b, err := ioutil.ReadAll(request.Body)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	defer request.Body.Close()
	var post models.Post
	err = json.Unmarshal(b, &post)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	var id = mux.Vars(request)["id"]
	db := common.GetDB()
	var result models.Post
	result.Id = -1
	rows, err := db.Query(`SELECT * FROM posts WHERE id = $1`, id)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	for rows.Next()  {
		err = rows.Scan(&result.Id, &result.Author, &result.Created, &result.Forum,
			&result.IsEdited, &result.Message, &result.Parent, &result.Thread, &result.Path)
	}
	if result.Id == -1 {
		var msg models.ResponseMessage
		msg.Message = `Can't find post post by id: ` + id
		output, err := json.Marshal(msg)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		w.Header().Set("content-type", "application/json")
		w.WriteHeader(404)
		w.Write(output)
		return
	} else {
		if post.Message != "" && post.Message != result.Message {
			err = db.QueryRow(`UPDATE posts SET message = $1, isEdited = true WHERE id = $2 RETURNING isEdited`, post.Message, id).Scan(&result.IsEdited)
			result.Message = post.Message
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
}