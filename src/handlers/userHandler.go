package handlers

import (
	"net/http"
	"io/ioutil"
	"../models"
	"encoding/json"
	"github.com/gorilla/mux"
	"../getters"
	"../common"
	"database/sql"
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
	exists, gotUsers := getters.GetUserByNickOrEmail(user.Nickname, user.Email)
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

func GetUser(w http.ResponseWriter, request *http.Request) {
	var user models.User
	user.Nickname = mux.Vars(request)["nick"]
	exists, gotUsers := getters.GetUserByNickOrEmail(user.Nickname, user.Email)
	if exists {
		user = gotUsers[0]
		output, err := json.Marshal(user)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		w.Header().Set("content-type", "application/json")
		w.WriteHeader(200)
		w.Write(output)
		return
	}
	var Response models.ResponseMessage
	Response.Message = "Can't find user by nickname: " + user.Nickname
	output, err := json.Marshal(Response)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(404)
	w.Write(output)

}

func UpdateUser(w http.ResponseWriter, request *http.Request) {
	b, err := ioutil.ReadAll(request.Body)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	defer request.Body.Close()
	exists, gotUsers := getters.GetUserByNickOrEmail(mux.Vars(request)["nick"], "")
	if !exists {
		var errorMsg models.ResponseMessage
		errorMsg.Message = "Can't find user by nickname: " + mux.Vars(request)["nick"]
		output, err := json.Marshal(errorMsg)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		w.Header().Set("content-type", "application/json")
		w.WriteHeader(404)
		w.Write(output)
		return
	}

	var user models.User
	err = json.Unmarshal(b, &user)
	user.Nickname = mux.Vars(request)["nick"]
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	var conflictNick = getters.GetNickByEmail(user.Email)
	if conflictNick != "" {
		var errorMsg models.ResponseMessage
		errorMsg.Message = "This email is already registered by user: " + conflictNick
		output, err := json.Marshal(errorMsg)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		w.Header().Set("content-type", "application/json")
		w.WriteHeader(409)
		w.Write(output)
		return
	}

	if user.Email == "" {
		user.Email = gotUsers[0].Email
	}
	if user.Fullname == "" {
		user.Fullname = gotUsers[0].Fullname
	}
	if user.About == "" {
		user.About = gotUsers[0].About
	}

	db := common.GetDB()
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(200)

	_, err = db.Exec(`UPDATE users SET about = $1, email = $2, fullname = $3 WHERE nickname = $4`, user.About, user.Email, user.Fullname, user.Nickname)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	output, err := json.Marshal(user)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.Write(output)
}

func GetThreadUsers(w http.ResponseWriter, request *http.Request) {
	var slug = mux.Vars(request)["slug"]
	limit := request.URL.Query().Get("limit")
	since := request.URL.Query().Get("since")
	desc := request.URL.Query().Get("desc")

	db := common.GetDB()
	var rows *sql.Rows
	var err error
	if desc == "false"  || desc == "" {
		if since == "" {
			if limit == "" {
				rows, err = db.Query("SELECT DISTINCT u.* FROM users AS u JOIN threads AS t ON u.nickname = t.author " +
					"JOIN posts AS p ON u.nickname = p.author WHERE t.forum = $1 AND p.forum = $1 ORDER BY u.nickname", slug)
			} else {
				rows, err = db.Query("SELECT DISTINCT u.* FROM users AS u JOIN threads AS t ON u.nickname = t.author " +
					"JOIN posts AS p ON u.nickname = p.author WHERE t.forum = $1 AND p.forum = $1 ORDER BY u.nickname LIMIT $2", slug, limit)
			}
		} else {
			if limit == "" {
				rows, err = db.Query("SELECT DISTINCT u.* FROM users AS u JOIN threads AS t ON u.nickname = t.author " +
					"JOIN posts AS p ON u.nickname = p.author WHERE t.forum = $1 AND p.forum = $1 AND u.nickname > $2 ORDER BY u.nickname", slug, since)
			} else {
				rows, err = db.Query("SELECT DISTINCT u.* FROM users AS u JOIN threads AS t ON u.nickname = t.author " +
					"JOIN posts AS p ON u.nickname = p.author WHERE t.forum = $1 AND p.forum = $1 AND u.nickname > $3 ORDER BY u.nickname LIMIT $2", slug, limit, since)
			}
		}
	} else {
		if since == "" {
			if limit == "" {
				rows, err = db.Query("SELECT DISTINCT u.* FROM users AS u JOIN threads AS t ON u.nickname = t.author " +
					"JOIN posts AS p ON u.nickname = p.author WHERE t.forum = $1 AND p.forum = $1 ORDER BY u.nickname DESC", slug)
			} else {
				rows, err = db.Query("SELECT DISTINCT u.* FROM users AS u JOIN threads AS t ON u.nickname = t.author " +
					"JOIN posts AS p ON u.nickname = p.author WHERE t.forum = $1 AND p.forum = $1 ORDER BY u.nickname DESC LIMIT $2", slug, limit)
			}
		} else {
			if limit == "" {
				rows, err = db.Query("SELECT DISTINCT u.* FROM users AS u JOIN threads AS t ON u.nickname = t.author " +
					"JOIN posts AS p ON u.nickname = p.author WHERE t.forum = $1 AND p.forum = $1 AND u.nickname < $2 ORDER BY u.nickname DESC", slug, since)
			} else {
				rows, err = db.Query("SELECT DISTINCT u.* FROM users AS u JOIN threads AS t ON u.nickname = t.author " +
					"JOIN posts AS p ON u.nickname = p.author WHERE t.forum = $1 AND p.forum = $1 AND u.nickname < $3 ORDER BY u.nickname DESC LIMIT $2", slug, limit, since)
			}
		}
	}
	//rows, err := db.Query("SELECT DISTINCT u.* FROM users AS u JOIN threads AS t ON u.nickname = t.author " +
	//	"JOIN posts AS p ON u.nickname = p.author WHERE t.forum = $1 AND p.forum = $1 ORDER BY u.nickname LIMIT $2", slug, limit)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	users := make([]models.User, 0)
	for rows.Next() {
		var gotUser models.User
		err = rows.Scan(&gotUser.About, &gotUser.Email, &gotUser.Fullname, &gotUser.Nickname, &gotUser.ID)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		users = append(users, gotUser)
	}
	output, err := json.Marshal(users)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(200)
	w.Write(output)
}