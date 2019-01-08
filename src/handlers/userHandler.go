package handlers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx"

	"github.com/igor-dyrov/forum-db/src/common"
	"github.com/igor-dyrov/forum-db/src/getters"
	"github.com/igor-dyrov/forum-db/src/models"
)

func CreateUser(w http.ResponseWriter, request *http.Request) {

	body, err := ioutil.ReadAll(request.Body)
	PanicIfError(err)
	defer request.Body.Close()

	var user models.User

	err = json.Unmarshal(body, &user)
	PanicIfError(err)

	user.Nickname = mux.Vars(request)["nick"]

	exists, gotUsers := getters.GetUserByNickOrEmail(user.Nickname, user.Email)

	if exists {
		WriteResponce(w, 409, gotUsers)
		return
	}

	conn := common.GetConnection()
	defer common.Release(conn)

	_, err = conn.Exec("INSERT INTO users (about, email, fullname, nickname) VALUES ($1, $2, $3, $4)", user.About, user.Email, user.Fullname, user.Nickname)
	PanicIfError(err)

	WriteResponce(w, 201, user)
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

func GetForumUsers2(w http.ResponseWriter, request *http.Request) {

	var slug = mux.Vars(request)["slug"]

	limit := request.URL.Query().Get("limit")
	since := request.URL.Query().Get("since")
	desc := request.URL.Query().Get("desc")

	if getters.GetForumBySlug(slug).Slug == "" {
		WriteNotFoundMessage(w, "Can`t find forum with slug: "+slug)
		return
	}

	db := common.GetPool()
	var rows *pgx.Rows
	var err error
	if desc == "false" || desc == "" {
		if since == "" {
			if limit == "" {
				rows, err = db.Query("SELECT DISTINCT u.* FROM users AS u JOIN posts AS p ON u.nickname = p.author WHERE p.forum = $1 UNION "+
					"SELECT DISTINCT u.* FROM users AS u JOIN threads AS t ON u.nickname = t.author WHERE t.forum = $1 ORDER BY nickname ASC", slug)
			} else {
				rows, err = db.Query("SELECT DISTINCT u.* FROM users AS u JOIN posts AS p ON u.nickname = p.author WHERE p.forum = $1 UNION "+
					"SELECT DISTINCT u.* FROM users AS u JOIN threads AS t ON u.nickname = t.author WHERE t.forum = $1 ORDER BY nickname ASC LIMIT $2", slug, limit)
			}
		} else {
			if limit == "" {
				rows, err = db.Query("SELECT DISTINCT u.* FROM users AS u JOIN posts AS p ON u.nickname = p.author WHERE p.forum = $1 AND u.nickname > $2 UNION "+
					"SELECT DISTINCT u.* FROM users AS u JOIN threads AS t ON u.nickname = t.author WHERE t.forum = $1 AND u.nickname > $2 ORDER BY nickname ASC", slug, since)
			} else {
				rows, err = db.Query("SELECT DISTINCT u.* FROM users AS u JOIN posts AS p ON u.nickname = p.author WHERE p.forum = $1 AND u.nickname > $2 UNION "+
					"SELECT DISTINCT u.* FROM users AS u JOIN threads AS t ON u.nickname = t.author WHERE t.forum = $1 AND u.nickname > $2 ORDER BY nickname ASC LIMIT $3", slug, since, limit)
			}
		}
	} else if desc == "true" {
		if since == "" {
			if limit == "" {
				rows, err = db.Query("SELECT DISTINCT u.* FROM users AS u JOIN posts AS p ON u.nickname = p.author WHERE p.forum = $1 UNION "+
					"SELECT DISTINCT u.* FROM users AS u JOIN threads AS t ON u.nickname = t.author WHERE t.forum = $1 ORDER BY nickname DESC", slug)
			} else {
				rows, err = db.Query("SELECT DISTINCT u.* FROM users AS u JOIN posts AS p ON u.nickname = p.author WHERE p.forum = $1 UNION "+
					"SELECT DISTINCT u.* FROM users AS u JOIN threads AS t ON u.nickname = t.author WHERE t.forum = $1 ORDER BY nickname DESC LIMIT $2", slug, limit)
			}
		} else {
			if limit == "" {
				rows, err = db.Query("SELECT DISTINCT u.* FROM users AS u JOIN posts AS p ON u.nickname = p.author WHERE p.forum = $1 AND u.nickname < $2 UNION "+
					"SELECT DISTINCT u.* FROM users AS u JOIN threads AS t ON u.nickname = t.author WHERE t.forum = $1 AND u.nickname < $2 ORDER BY nickname DESC", slug, since)
			} else {
				rows, err = db.Query("SELECT DISTINCT u.* FROM users AS u JOIN posts AS p ON u.nickname = p.author WHERE p.forum = $1 AND u.nickname < $2 UNION "+
					"SELECT DISTINCT u.* FROM users AS u JOIN threads AS t ON u.nickname = t.author WHERE t.forum = $1 AND u.nickname < $2 ORDER BY nickname DESC LIMIT $3", slug, since, limit)
			}
		}
	}
	defer rows.Close()
	PanicIfError(err)

	users := make([]models.User, 0)

	for rows.Next() {
		var gotUser models.User
		PanicIfError(rows.Scan(&gotUser.About, &gotUser.Email, &gotUser.Fullname, &gotUser.Nickname, &gotUser.ID))
		users = append(users, gotUser)
	}

	WriteResponce(w, 200, users)
}

func GetForumUsers(w http.ResponseWriter, request *http.Request) {

	var slug = mux.Vars(request)["slug"]

	limit := request.URL.Query().Get("limit")
	since := request.URL.Query().Get("since")
	descStr := request.URL.Query().Get("desc")
	desc := descStr == "true"

	forum := getters.GetForumBySlug(slug)
	if forum.Slug == "" {
		WriteNotFoundMessage(w, "Can`t find forum with slug: "+slug)
		return
	}

	sinceStr := ""
	if since != "" {
		if desc {
			sinceStr = " AND uf.username < '" + since + "'"
		} else {
			sinceStr = " AND uf.username > '" + since + "'"
		}
	}

	order := " ASC"
	if desc {
		order = " DESC"
	}

	limitStr := ""
	if limit != "" {
		limitStr = " LIMIT " + limit
	}

	query := fmt.Sprintf(
		"SELECT nickname, fullname, about, email FROM users u JOIN forum_users uf ON u.nickname = uf.username"+
			" WHERE uf.forum = '%s' %s ORDER BY uf.username %s %s;",
		forum.Slug, sinceStr, order, limitStr,
	)

	rows, err := common.GetPool().Query(query)
	defer rows.Close()
	PanicIfError(err)

	users := make([]models.User, 0)

	for rows.Next() {
		var user models.User
		PanicIfError(rows.Scan(
			&user.Nickname,
			&user.Fullname,
			&user.About,
			&user.Email,
		))

		users = append(users, user)
	}

	WriteResponce(w, 200, users)
}
