package handlers

import (
	"net/http"
	"io/ioutil"
	"../models"
	"encoding/json"
	"github.com/gorilla/mux"
	"../getters"
	"../common"
)

func CreateUser(w http.ResponseWriter, request *http.Request) {
	b, err := ioutil.ReadAll(request.Body)

	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	defer request.Body.Close()

	var user models.User

	err = json.Unmarshal(b, &user)

	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	user.Nickname = mux.Vars(request)["nick"]

	db := common.GetDB()

	exists, gotUsers := getters.GetUserByNick(user)

	if exists {
		output, err := json.Marshal(gotUsers)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		w.WriteHeader(409)
		w.Header().Set("content-type", "application/json")
		w.Write(output)
		return
	}

	_, err = db.Exec("INSERT INTO users (about, email, fullname, nickname) VALUES ($1, $2, $3, $4)", user.About, user.Email, user.Fullname, user.Nickname)

	output, err := json.Marshal(user)

	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	w.Header().Set("content-type", "application/json")
	w.WriteHeader(201)
	w.Write(output)
}




