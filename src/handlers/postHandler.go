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

func uniqueStrings(strArray []string) (result map[string]bool) {
	for _, s := range strArray {
		result[s] = true
	}
	return result
}

func uniqueIDs(idArray []int32) (result map[int32]bool) {
	for _, s := range idArray {
		result[s] = true
	}
	return result
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

	threadExists, threadID, forumSlug := getters.GetThreadIDAndForumBySlugOrID(ThreadSlugOrID)

	if !threadExists {
		WriteNotFoundMessage(w, "Can't find thread by id: "+ThreadSlugOrID)
		return
	}

	if len(posts) == 0 {
		WriteResponce(w, 201, posts)
		return
	}

	uniqueParentIDs := make(map[int32]bool)
	uniqueAuthors := make(map[string]bool)

	for i := range posts {
		posts[i].Forum = forumSlug
		posts[i].Thread = threadID
		posts[i].Created = curTime

		uniqueAuthors[posts[i].Author] = true
		if posts[i].Parent != 0 {
			uniqueParentIDs[posts[i].Parent] = true
		}
	}

	existedUsers := getters.GetUsersByNicknames(uniqueAuthors)
	if len(existedUsers) != len(uniqueAuthors) {
		WriteNotFoundMessage(w, "Can't find authors")
		return

	}

	if len(uniqueParentIDs) != 0 && len(getters.GetPostsIDByIDs(uniqueParentIDs, threadID)) != len(uniqueParentIDs) {
		WriteResponce(w, 409, models.ResponseMessage{Message: "Parent post was created in another thread"})
		return

	}

	pool := common.GetPool()

	for i := range posts {

		err := pool.QueryRow(
			"INSERT INTO posts (author, created, forum, message, thread, parent, path) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id;",
			posts[i].Author, posts[i].Created, posts[i].Forum,
			posts[i].Message, posts[i].Thread, posts[i].Parent,
			pq.Array(posts[i].Path),
		).Scan(
			&posts[i].Id,
		)
		PanicIfError(err)
	}

	_, err = pool.Exec("UPDATE forums SET posts = posts + $1 WHERE slug = $2", len(posts), forumSlug)
	PanicIfError(err)

	WriteResponce(w, 201, posts)
}

func GetPost(w http.ResponseWriter, request *http.Request) {

	id, err := strconv.ParseInt(mux.Vars(request)["id"], 10, 32)
	PanicIfError(err)
	postID := int32(id)

	isPostExists, post := getters.GetPostByID(postID)
	if !isPostExists {
		WriteNotFoundMessage(w, "Can't find post post by id: "+mux.Vars(request)["id"])
		return
	}

	related := request.URL.Query().Get("related")
	additions := strings.Split(related, ",")

	PostInfo := new(models.PostDetails)

	PostInfo.Post = &post

	for _, info := range additions {
		if info == "user" {
			var tempUser models.User

			_, tempUser = getters.GetUserByNickname(PostInfo.Post.Author)
			PostInfo.Author = &tempUser
		}
		if info == "thread" {
			var tempThread models.Thread
			_, tempThread = getters.GetThreadById(PostInfo.Post.Thread)
			PostInfo.Thread = &tempThread
		}
		if info == "forum" {
			var tempForum models.Forum

			tempForum = getters.GetForumBySlug(PostInfo.Post.Forum)
			PostInfo.Forum = &tempForum
		}
	}

	WriteResponce(w, 200, PostInfo)
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
