package handlers

import (
	"strconv"
	"strings"
	"time"

	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/lib/pq"

	"github.com/igor-dyrov/forum-db/src/common"
	"github.com/igor-dyrov/forum-db/src/getters"
	"github.com/igor-dyrov/forum-db/src/models"
)

func requiredParents(posts []models.Post) []int {

	parents := make(map[int]bool)
	requiredParents := make([]int, 0, len(posts))

	for i := 0; i < len(posts); i++ {
		parents[posts[i].Parent] = true
	}

	parents[0] = false

	for id, isRequired := range parents {
		if isRequired {
			requiredParents = append(requiredParents, id)
		}
	}

	return requiredParents
}

func CreatePosts(w http.ResponseWriter, request *http.Request) {

	curTime := time.Now().Truncate(time.Millisecond).UTC()

	body, err := ioutil.ReadAll(request.Body)
	PanicIfError(err)
	defer request.Body.Close()

	var posts []models.Post
	err = json.Unmarshal(body, &posts)
	PanicIfError(err)

	var ThreadSlugOrID = mux.Vars(request)["slug_or_id"]

	id, err := strconv.Atoi(ThreadSlugOrID)
	var forum string
	if err == nil { //got id
		if !getters.ThreadExists(id) {
			WriteNotFoundMessage(w, "Can't find post thread by id: "+ThreadSlugOrID)
			return
		}
		forum = getters.GetSlugById(id)
	}

	err = nil
	var Thread int
	if forum != "" {
		Thread = id
	} else {
		forum = getters.GetThreadSlug(ThreadSlugOrID) //get forum from thread by slug
		Thread = getters.GetThreadId(ThreadSlugOrID)
		if Thread == -1 {
			WriteNotFoundMessage(w, "Can't find post thread by slug: "+ThreadSlugOrID)
			return
		}
	}

	if len(posts) == 0 {
		WriteResponce(w, 201, posts)
		return
	}

	db := common.GetDB()

	var maxId int = 0
	err = db.QueryRow(`SELECT MAX(id) FROM posts`).Scan(&maxId)
	err = nil
	maxId++

	for i := range posts {
		posts[i].Forum = forum
		posts[i].Thread = Thread
		posts[i].Created = curTime

		if !getters.UserExists(posts[i].Author) {
			WriteNotFoundMessage(w, "Can't find thread author by nickname: "+posts[i].Author)
			return
		}
		if posts[i].Parent != 0 && !getters.CheckParent(posts[i].Parent, posts[i].Thread) {
			WriteResponce(w, 409, models.ResponseMessage{Message: "Parent post was created in another thread"})
			return
		}

		var parentPath = getters.GetPathById(posts[i].Parent)
		for j := range parentPath {
			posts[i].Path = append(posts[i].Path, parentPath[j])
		}

		posts[i].Path = append(posts[i].Path, maxId+i)
		db.QueryRow(`INSERT INTO posts (author, created, forum, message, thread, parent, path) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`,
			posts[i].Author, posts[i].Created, posts[i].Forum, posts[i].Message, posts[i].Thread, posts[i].Parent, pq.Array(posts[i].Path)).Scan(&posts[i].Id)
		PanicIfError(err)
	}

	_, err = db.Exec("UPDATE forums SET posts = posts + $1 WHERE slug = $2", len(posts), forum)
	PanicIfError(err)

	WriteResponce(w, 201, posts)
}

func GetPost(w http.ResponseWriter, request *http.Request) {
	var id = mux.Vars(request)["id"]
	related := request.URL.Query().Get("related")
	additions := strings.Split(related, ",")
	PostInfo := new(models.PostDetails)
	db := common.GetDB()
	var result models.Post
	result.Id = -1
	rows, err := db.Query(`SELECT * FROM posts WHERE id = $1`, id)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	var gotPath string
	for rows.Next() {
		err = rows.Scan(&result.Id, &result.Author, &result.Created, &result.Forum,
			&result.IsEdited, &result.Message, &result.Parent, &result.Thread, &gotPath)
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
	}
	//if len(gotPath) > 2 {
	IDs := strings.Split(gotPath[1:len(gotPath)-1], ",")
	for index := range IDs {
		item, _ := strconv.Atoi(IDs[index])
		result.Path = append(result.Path, item)
	}
	//}
	PostInfo.Post = &result
	var tempUser models.User
	var tempThread models.Thread
	var tempForum models.Forum
	for _, info := range additions {
		if info == "user" {
			tempUser = getters.GetUserByNick(result.Author)
			PostInfo.Author = &tempUser
		}
		if info == "thread" {
			tempThread = getters.GetThreadById(PostInfo.Post.Thread)
			PostInfo.Thread = &tempThread
		}
		if info == "forum" {
			tempForum = getters.GetForumBySlug(PostInfo.Post.Forum)
			PostInfo.Forum = &tempForum
		}
	}
	var output []byte
	output, err = json.Marshal(PostInfo)
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
	for rows.Next() {
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
