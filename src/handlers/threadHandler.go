package handlers

import (
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

	if !getters.CheckUserByNickname(thread.Author) {
		WriteNotFoundMessage(w, "Can't find thread author by nickname: "+thread.Author)
		return
	}

	thread.Forum = getters.GetForumSlug(thread.Forum)
	if thread.Forum == "" {
		WriteNotFoundMessage(w, "Can't find thread forum by slug: "+thread.Forum)
		return
	}

	if gotThread := getters.ConnGetThreadBySlug(thread.Slug); gotThread != nil {
		WriteResponce(w, 409, gotThread)
		return
	}

	conn := common.GetPool()

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

		*posts = append(*posts, result)
	}
}

func GetThreadPosts(w http.ResponseWriter, request *http.Request) {

	var slug_or_id = mux.Vars(request)["slug_or_id"]

	limit := request.URL.Query().Get("limit")
	since := request.URL.Query().Get("since")
	sort := request.URL.Query().Get("sort")
	desc := request.URL.Query().Get("desc")

	// fmt.Printf("%v | sort=%v limit=%v since=%v desc=%v\n", request.URL.Path, sort, limit, since, desc)

	thread := getters.GetThreadBySlugOrID(slug_or_id)
	if thread == nil {
		WriteNotFoundMessage(w, "Can't find post thread by id: "+slug_or_id)
		return
	}

	id := thread.ID

	if sort == "flat" || sort == "" {
		WriteResponce(w, 200, getters.GetPostsFlatSort(id, limit, since, desc == "true"))
		return
	}

	if sort == "tree" {
		WriteResponce(w, 200, getters.GetPostsTreeSort(id, limit, since, desc == "true"))
		return
	}

	if sort == "parent_tree" {
		WriteResponce(w, 200, getters.GetPostsParentTreeSort(id, limit, since, desc == "true"))
		return
	}
}
