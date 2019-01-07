package handlers

import (
	"fmt"
	"strconv"

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
	body, err := ioutil.ReadAll(request.Body)
	defer request.Body.Close()
	PanicIfError(err)

	var thread models.Thread
	err = json.Unmarshal(body, &thread)
	PanicIfError(err)

	thread.Forum = mux.Vars(request)["slug"]

	conn := common.GetConnection()
	defer common.Release(conn)

	if !getters.CheckUserByNickname(thread.Author, conn) {
		WriteNotFoundMessage(w, "Can't find thread author by nickname: "+thread.Author)
		return
	}

	thread.Forum = getters.GetForumSlug(thread.Forum, conn)
	if thread.Forum == "" {
		WriteNotFoundMessage(w, "Can't find thread forum by slug: "+thread.Forum)
		return
	}

	if gotThread := getters.ConnGetThreadBySlug(thread.Slug, conn); gotThread != nil {
		WriteResponce(w, 409, gotThread)
		return
	}

	if thread.Slug != "" {
		err := conn.QueryRow("INSERT INTO threads (slug, created, message, title, author, forum, votes) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id",
			thread.Slug, thread.Created, thread.Message, thread.Title, thread.Author, thread.Forum, thread.Votes).Scan(&thread.ID)
		PanicIfError(err)
	} else {
		err := conn.QueryRow("INSERT INTO threads (created, message, title, author, forum, votes) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id",
			thread.Created, thread.Message, thread.Title, thread.Author, thread.Forum, thread.Votes).Scan(&thread.ID)
		PanicIfError(err)
	}

	_, err = conn.Exec("UPDATE forums SET threads = threads + 1 WHERE slug = $1", thread.Forum)
	PanicIfError(err)

	WriteResponce(w, 201, thread)
}

func GetThreads(w http.ResponseWriter, request *http.Request) {

	var slug = mux.Vars(request)["slug"]

	limit := request.URL.Query().Get("limit")
	since := request.URL.Query().Get("since")
	desc := request.URL.Query().Get("desc")

	if !getters.ForumSlugExists(slug) {
		WriteNotFoundMessage(w, "Can't find forum by slug: "+slug)
		return
	}

	gotThreads := getters.GetThreads(slug, limit, since, desc)

	WriteResponce(w, 200, gotThreads)
}

func ThreadDetails(w http.ResponseWriter, request *http.Request) {

	var slugOrID = mux.Vars(request)["slug_or_id"]

	thread := getters.GetThreadBySlugOrID(slugOrID)
	if thread == nil {
		WriteNotFoundMessage(w, "Can't find thread by slug-or-id: "+slugOrID)
		return
	}

	WriteResponce(w, 200, thread)
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
		PanicIfError(rows.Scan(&result.Id, &result.Author, &result.Created, &result.Forum,
			&result.IsEdited, &result.Message, &result.Parent, &result.Thread, &gotPath))

		// IDs := strings.Split(gotPath[1:len(gotPath)-1], ",")
		// for index := range IDs {
		// 	item, _ := strconv.ParseInt(IDs[index], 10, 32)
		// 	result.Path = append(result.Path, int32(item))
		// }
		*posts = append(*posts, result)
	}
}

// func getPostsByQuery(query string, params ...interface{}) []models.Post {
// 	postArr := make([]models.Post, 0)

// 	rows, err := common.GetPool().Query(query, params...)
// 	defer rows.Close()
// 	PanicIfError(err)

// 	for rows.Next() {
// 		var post model.Post
// 	}
// }

func GetThreadPosts(w http.ResponseWriter, request *http.Request) {

	var slug_or_id = mux.Vars(request)["slug_or_id"]

	limit := request.URL.Query().Get("limit")
	since := request.URL.Query().Get("since")
	sort := request.URL.Query().Get("sort")
	desc := request.URL.Query().Get("desc")

	// fmt.Printf("%v | sort=%v limit=%v since=%v desc=%v\n", request.URL.Path, sort, limit, since, desc)

	db := common.GetDB()

	thread := getters.GetThreadBySlugOrID(slug_or_id)
	if thread == nil {
		WriteNotFoundMessage(w, "Can't find post thread by id: "+slug_or_id)
		return
	}

	id := thread.ID
	var err error

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
		defer rows.Close()
		PanicIfError(err)

		HandlePostRows(rows, &posts)
		WriteResponce(w, 200, posts)
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
					// req += `ORDER BY path DESC LIMIT ` + limit
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

		PanicIfError(err)
		defer rows.Close()
		HandlePostRows(rows, &posts)

	} else if sort == "parent_tree" {
		parentPosts := getters.GetParentPosts(id)

		fmt.Printf("%v | %v limit=%v since=%v desc=%v -> len(parents)=%d\n", request.URL.Path, sort, limit, since, desc, len(parentPosts))

		pageSize, _ := strconv.Atoi(limit)
		if desc == "false" || desc == "" {
			if since == "" {
				for i := range parentPosts {
					if i < pageSize {
						rows, err := db.Query("SELECT * FROM posts WHERE thread = $1 AND path[1] = $2 ORDER BY path ASC, id ASC", id, parentPosts[i].Id)
						defer rows.Close()
						PanicIfError(err)
						HandlePostRows(rows, &posts)
					}
				}
			} else {
				for i := range parentPosts {
					if i <= pageSize {
						rows, err := db.Query("SELECT p1.* FROM posts AS p1 JOIN posts AS p2 ON p1.thread = $1 AND p1.path[1] > p2.path[1]"+
							" AND p2.id = $2 WHERE p1.path[1] = $3 ORDER BY p1.path ASC, p1.id ASC", id, since, parentPosts[i].Id)
						defer rows.Close()
						PanicIfError(err)
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
						defer rows.Close()
						PanicIfError(err)
						HandlePostRows(rows, &posts)
					}
				}
			} else {
				for i := range parentPosts {
					if i <= pageSize {
						var lastParent = parentPosts[len(parentPosts)-1-i].Id
						rows, err := db.Query("SELECT p1.* FROM posts AS p1 JOIN posts AS p2 ON p1.thread = $1 AND p1.path[1] < p2.path[1]"+
							" AND p2.id = $2 WHERE p1.path[1] = $3 ORDER BY p1.path ASC, p1.id ASC", id, since, lastParent)
						defer rows.Close()
						PanicIfError(err)
						HandlePostRows(rows, &posts)
					}
				}
			}
		}
	}

	WriteResponce(w, 200, posts)
}
