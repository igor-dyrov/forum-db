package getters

import (
	_ "github.com/lib/pq"
	"../models"
	"../common"
)

func GetUserByNickOrEmail(nickname string, email string) (bool, []models.User) {
	db := common.GetDB()

	rows, err := db.Query(`SELECT * from users WHERE email = $1 OR nickname = $2`, email, nickname)
	result := make([]models.User, 0)

	if err != nil {
		if err != nil {
			return false, result
		}
	}

	for rows.Next() {
		var gotUser models.User
		err = rows.Scan(&gotUser.About, &gotUser.Email, &gotUser.Fullname, &gotUser.Nickname, &gotUser.ID)
		if gotUser.Nickname != "" {
			result = append(result, gotUser)
		}
	}

	if err == nil && len(result) != 0 {
		return true, result
	}
	return false, result
}

func GetNickByEmail(email string) string {
	db := common.GetDB()
	rows, err := db.Query("SELECT nickname from users WHERE email = $1", email)
	if err != nil {
		return ""
	}
	var nickname string
	for rows.Next() {
		err = rows.Scan(&nickname)
		if err != nil {
			return ""
		}
	}
	return nickname
}

func UserExists(nickname string) bool {
	db := common.GetDB()
	rows, err := db.Query("SELECT * from users WHERE nickname = $1", nickname)
	if err != nil {
		return false
	}
	for rows.Next() {
		return true
	}
	return false
}

func GetIdByNickname(nickname string) int {
	db := common.GetDB()
	rows, err := db.Query("SELECT id FROM users WHERE nickname = $1", nickname)
	if err != nil {
		return 0
	}
	var id int
	for rows.Next() {
		err = rows.Scan(&id)
		if err != nil {
			return 0
		}
	}
	return id
}

func GetForumBySlug(slug string) models.Forum {
	db := common.GetDB()
	rows, err := db.Query("SELECT * from forums WHERE slug = $1", slug)
	var result models.Forum
	if err != nil {
		return models.Forum{}
	}
	for rows.Next() {
		err = rows.Scan(&result.ID, &result.Slug, &result.Title, &result.Author, &result.Threads, &result.Posts)
	}
	if err != nil {
		return models.Forum{}
	}
	return result
}