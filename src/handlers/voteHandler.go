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

// CreateVote ...
func CreateVote(w http.ResponseWriter, request *http.Request) {
	w.Header().Set("content-type", "application/json")

	body, err := ioutil.ReadAll(request.Body)
	defer request.Body.Close()
	PanicIfError(err)

	var vote models.Vote
	err = json.Unmarshal(body, &vote)
	PanicIfError(err)

	conn := common.GetConnection()
	defer common.Release(conn)

	if !getters.CheckUserByNickname(vote.Nickname, conn) {
		WriteNotFoundMessage(w, "Can't find thread author by nickname: "+vote.Nickname)
		return
	}

	var slugOrID = mux.Vars(request)["slug_or_id"]
	thread := getters.GetThreadBySlugOrID(slugOrID)
	if thread == nil {
		WriteNotFoundMessage(w, "Can't find post thread by id: "+slugOrID)
		return
	}

	numOfVoices := 0
	oldVote := getters.GetVote(vote.Nickname, thread.ID)

	db := common.GetDB()

	if oldVote.ID != -1 {
		if oldVote.Voice != vote.Voice {
			_, err := db.Exec("UPDATE votes SET voice = $1 WHERE id = $2;", vote.Voice, oldVote.ID)
			PanicIfError(err)
			numOfVoices = vote.Voice * 2
		} else {
			// log.Printf("Thread: %v", *thread)
			WriteResponce(w, 200, thread)
			return
		}
	} else {
		_, err := db.Exec("INSERT INTO votes (nickname, voice, thread) VALUES ($1, $2, $3);", vote.Nickname, vote.Voice, thread.ID)
		PanicIfError(err)
		numOfVoices = vote.Voice
	}

	rows, err := db.Query("UPDATE threads SET votes = votes + $1 WHERE id = $2 RETURNING votes;", numOfVoices, thread.ID)
	defer rows.Close()
	PanicIfError(err)

	if rows.Next() {
		err := rows.Scan(&thread.Votes)
		PanicIfError(err)
	}

	// log.Printf("ThreadID: %d, user: %s, voice: %d -> %d", thread.ID, vote.Nickname, vote.Voice, thread.Votes)

	WriteResponce(w, 200, thread)
}
