package handlers

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/igor-dyrov/forum-db/src/common"
	"github.com/igor-dyrov/forum-db/src/getters"
	"github.com/igor-dyrov/forum-db/src/models"
)

func CreateForum(w http.ResponseWriter, request *http.Request) {
	b, err := ioutil.ReadAll(request.Body)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	defer request.Body.Close()
	var forum models.Forum
	err = json.Unmarshal(b, &forum)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	if !getters.UserExists(forum.Author) {
		var message models.ResponseMessage
		message.Message = "Can't find user with nickname: " + forum.Author
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

	var gotForum = getters.GetForumBySlug(forum.Slug)
	if gotForum.Author != "" {
		output, err := json.Marshal(gotForum)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		w.Header().Set("content-type", "application/json")
		w.WriteHeader(409)
		w.Write(output)
		return
	}

	db := common.GetDB()
	_, author := getters.GetUserByNickOrEmail(forum.Author, "")
	forum.Author = author[0].Nickname
	_, err = db.Exec(`INSERT INTO forums (slug, title, author) VALUES ($1, $2, $3)`, forum.Slug, forum.Title, forum.Author)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	output, err := json.Marshal(forum)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	w.Header().Set("content-type", "application/json")
	w.WriteHeader(201)
	w.Write(output)
}

func GetForum(w http.ResponseWriter, request *http.Request) {

	slug := mux.Vars(request)["slug"]

	forum := getters.GetForumBySlug(slug)

	if forum.Author != "" {
		WriteResponce(w, 200, forum)
		return
	}

	WriteNotFoundMessage(w, "Can`t find forum with slug: "+slug)

}
