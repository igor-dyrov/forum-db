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

func CreateVote(w http.ResponseWriter, request *http.Request) {
	b, err := ioutil.ReadAll(request.Body)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	defer request.Body.Close()
	var vote models.Vote
	err = json.Unmarshal(b, &vote)
	if err != nil {
		http.Error(w, err.Error() + "24", 500)
		return
	}
	var slug_or_id = mux.Vars(request)["slug_or_id"]

	db := common.GetDB()
	var thread models.Thread
	id, err := strconv.Atoi(slug_or_id)
	if  err == nil {
		if !getters.ThreadExists(id) {
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
		thread.ID = id
	} else {
		thread.ID = getters.GetIdBySlug(slug_or_id)
		if thread.ID == -1 {
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
	var numOfVoices = 0
	oldVote := getters.GetVote(vote.Nickname, thread.ID)
	log.Println(`____________________________________`)
	log.Println(oldVote.ID)
	log.Println(vote.Nickname)
	log.Println(vote.Thread)
	log.Println(`____________________________________`)
	if oldVote.ID != -1 {
		if oldVote.Voice != vote.Voice {
			_, err = db.Query(`UPDATE votes SET voice = $1 WHERE id = $2`, vote.Voice, oldVote.ID)
			numOfVoices = vote.Voice * 2
		}
	} else {
		_, err = db.Query(`INSERT INTO votes (nickname, voice, thread) VALUES ($1, $2, $3)`, vote.Nickname, vote.Voice, thread.ID)
		numOfVoices = vote.Voice
	}
	if err != nil {
		http.Error(w, err.Error() + "66", 500)
		return
	}

	db.QueryRow(`UPDATE threads SET votes = votes + $1 WHERE id = $2 RETURNING *`, numOfVoices, thread.ID).Scan(&thread.ID, &thread.Slug, &thread.Created,
		&thread.Message, &thread.Title, &thread.Author, &thread.Forum, &thread.Votes)
	output, err := json.Marshal(thread)
	if err != nil {
		http.Error(w, err.Error() + "74", 500)
		return
	}
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(200)
	w.Write(output)
}