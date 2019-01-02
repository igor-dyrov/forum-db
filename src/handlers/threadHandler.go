package handlers

import (
	"strconv"
	"strings"

	"database/sql"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/igor-dyrov/forum-db/src/common"
	"github.com/igor-dyrov/forum-db/src/getters"
	"github.com/igor-dyrov/forum-db/src/models"
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
	if err == nil {
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

func UpdateThread(w http.ResponseWriter, request *http.Request) {
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
	slug_or_id := mux.Vars(request)["slug_or_id"]
	db := common.GetDB()
	id, err := strconv.Atoi(slug_or_id) //try to get id
	if err == nil {                     //got id
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
	} else { // got slug
		id = getters.GetIdBySlug(slug_or_id)
		if id == -1 {
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
	var req = `UPDATE threads SET `
	var result models.Thread
	if thread.Message != "" {
		req += `message = '` + thread.Message + `'`
		if thread.Title != "" {
			req += `, `
		}
	}
	if thread.Title != "" {
		req += `title = '` + thread.Title + `' `
	}
	req += `WHERE id = ` + strconv.Itoa(id)
	req += ` RETURNING *`
	if thread.Title == "" && thread.Message == "" {
		err = db.QueryRow(`SELECT * FROM threads WHERE id = $1`, id).Scan(&result.ID, &result.Slug, &result.Created, &result.Message,
			&result.Title, &result.Author, &result.Forum, &result.Votes)
	} else {
		err = db.QueryRow(req).Scan(&result.ID, &result.Slug, &result.Created, &result.Message,
			&result.Title, &result.Author, &result.Forum, &result.Votes)
	}
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
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

func HandlePostRows(rows *sql.Rows, posts *[]models.Post) {
	for rows.Next() {
		var result models.Post
		var gotPath string
		rows.Scan(&result.Id, &result.Author, &result.Created, &result.Forum,
			&result.IsEdited, &result.Message, &result.Parent, &result.Thread, &gotPath)
		IDs := strings.Split(gotPath[1:len(gotPath)-1], ",")
		for index := range IDs {
			item, _ := strconv.Atoi(IDs[index])
			result.Path = append(result.Path, item)
		}
		*posts = append(*posts, result)
	}
}

func GetThreadPosts(w http.ResponseWriter, request *http.Request) {
	var slug_or_id = mux.Vars(request)["slug_or_id"]
	limit := request.URL.Query().Get("limit")
	since := request.URL.Query().Get("since")
	sort := request.URL.Query().Get("sort")
	desc := request.URL.Query().Get("desc")

	db := common.GetDB()
	id, err := strconv.Atoi(slug_or_id) //try to get id
	if err == nil {                     //got id
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
	} else { // got slug
		id = getters.GetIdBySlug(slug_or_id)
		if id == -1 {
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

	var posts = make([]models.Post, 0)
	var req = `SELECT * FROM posts WHERE thread = ` + strconv.Itoa(id) + ` `
	if sort == "flat" || sort == "" {
		if limit != "" {
			if desc == "false" || desc == "" {
				if since != "" {
					req += `AND id >` + since + ` ORDER BY id ASC LIMIT ` + limit
				} else {
					req += `ORDER BY id LIMIT ` + limit
				}
			} else {
				if since != "" {
					req += `AND id <` + since + ` ORDER BY id DESC LIMIT ` + limit
				} else {
					req += `ORDER BY id DESC LIMIT ` + limit
				}
			}
		} else {
			if desc == "false" || desc == "" {
				if since != "" {
					req += `AND id >` + since + ` ORDER BY id ASC`
				} else {
					req += `ORDER BY id`
				}
			} else {
				if since != "" {
					req += `AND id <` + since + ` ORDER BY id DESC`
				} else {
					req += `ORDER BY id DESC`
				}
			}
		}
		rows, err := db.Query(req)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		HandlePostRows(rows, &posts)
		output, err := json.Marshal(posts)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		w.Header().Set("content-type", "application/json")
		w.WriteHeader(200)
		w.Write(output)
		return
	}
	var rows *sql.Rows
	if sort == "tree" {
		if limit != "" {
			if desc == "false" || desc == "" {
				if since != "" {
					rows, err = db.Query(`SELECT p1.* FROM posts AS p1 JOIN posts AS p2 ON p1.path > p2.path AND p2.id = $2 
		WHERE p1.thread = $1 ORDER BY path ASC LIMIT $3`, id, since, limit)
				} else {
					rows, err = db.Query(`SELECT * FROM posts WHERE thread = $1 ORDER BY path LIMIT $2`, id, limit)
				}
			} else {
				if since != "" {
					rows, err = db.Query(`SELECT p1.* FROM posts AS p1 JOIN posts AS p2 ON p1.path < p2.path AND p2.id = $2 
		WHERE p1.thread = $1 ORDER BY path DESC LIMIT $3`, id, since, limit)
				} else {
					rows, err = db.Query(`SELECT * FROM posts WHERE thread = $1 ORDER BY path DESC LIMIT $2`, id, limit)
					req += `ORDER BY path DESC LIMIT ` + limit
				}
			}
		} else {
			if desc == "false" || desc == "" {
				if since != "" {
					rows, err = db.Query(`SELECT p1.* FROM posts AS p1 JOIN posts AS p2 ON p1.path > p2.path AND p2.id = $2 
		WHERE p1.thread = $1 ORDER BY path ASC`, id, since)
				} else {
					rows, err = db.Query(`SELECT * FROM posts WHERE thread = $1 ORDER BY path ASC`, id)
				}
			} else {
				if since != "" {
					rows, err = db.Query(`SELECT p1.* FROM posts AS p1 JOIN posts AS p2 ON p1.path > p2.path AND p2.id = $2 
		WHERE p1.thread = $1 ORDER BY path DESC`, id, since)
				} else {
					rows, err = db.Query(`SELECT * FROM posts WHERE thread = $1 ORDER BY path DESC`, id)
				}
			}
		}
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		HandlePostRows(rows, &posts)
	} else if sort == "parent_tree" {
		parentPosts := getters.GetParentPosts(id)
		pageSize, _ := strconv.Atoi(limit)
		if desc == "false" || desc == "" {
			if since == "" {
				for i := range parentPosts {
					if i < pageSize {
						rows, err := db.Query("SELECT * FROM posts WHERE thread = $1 AND path[1] = $2 ORDER BY path ASC, id ASC", id, parentPosts[i].Id)
						if err != nil {
							http.Error(w, err.Error(), 500)
							return
						}
						HandlePostRows(rows, &posts)
					}
				}
			} else {
				for i := range parentPosts {
					if i <= pageSize {
						rows, err := db.Query("SELECT p1.* FROM posts AS p1 JOIN posts AS p2 ON p1.thread = $1 AND p1.path[1] > p2.path[1]"+
							" AND p2.id = $2 WHERE p1.path[1] = $3 ORDER BY p1.path ASC, p1.id ASC", id, since, parentPosts[i].Id)
						if err != nil {
							http.Error(w, err.Error(), 500)
							return
						}
						HandlePostRows(rows, &posts)
					}
				}
			}
		} else if desc == "true" {
			if since == "" {
				for i := range parentPosts {
					if i < pageSize {
						var lastParent = parentPosts[len(parentPosts)-1-i].Id
						rows, err := db.Query("SELECT * FROM posts WHERE thread = $1 AND path[1] = $2 ORDER BY path ASC, id ASC", id, lastParent)
						// log.Println(limit, since, sort, desc)
						// log.Println("here")
						if err != nil {
							http.Error(w, err.Error(), 500)
							return
						}
						HandlePostRows(rows, &posts)
					}
				}
			} else {
				for i := range parentPosts {
					if i <= pageSize {
						var lastParent = parentPosts[len(parentPosts)-1-i].Id
						rows, err := db.Query("SELECT p1.* FROM posts AS p1 JOIN posts AS p2 ON p1.thread = $1 AND p1.path[1] < p2.path[1]"+
							" AND p2.id = $2 WHERE p1.path[1] = $3 ORDER BY p1.path ASC, p1.id ASC", id, since, lastParent)
						if err != nil {
							http.Error(w, err.Error(), 500)
							return
						}
						HandlePostRows(rows, &posts)
					}
				}
			}
		}
	}
	output, err := json.Marshal(posts)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(200)
	w.Write(output)
}
