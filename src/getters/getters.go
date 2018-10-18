package getters

import (
	_ "github.com/lib/pq"
	"../models"
	"../common"
	"database/sql"
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
	rows, err := db.Query("SELECT email from users WHERE nickname = $1", nickname)
	if err != nil {
		return false
	}
	var email string
	for rows.Next() {
		rows.Scan(&email)
	}
	if err != nil || email == "" {
		return false
	}
	return true
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

func GetThreadBySlug(slug string) models.Thread {
	db := common.GetDB()
	rows, err := db.Query("SELECT * FROM threads WHERE slug = $1", slug)
	if err != nil {
		return models.Thread{}
	}
	var result models.Thread
	for rows.Next() {
		err = rows.Scan(&result.ID, &result.Slug, &result.Created, &result.Message, &result.Title, &result.Author, &result.Forum, &result.Votes)
	}
	if err != nil {
		return models.Thread{}
	}
	return result
}

func GetThreads(forum string, limit string, since string, desc string) ([]models.Thread) {
	db := common.GetDB()
	var rows *sql.Rows
	var err error

	if limit != "" {
		if since != "" {
			if desc == "true" {
				rows, err = db.Query("SELECT * FROM threads WHERE forum = $1 AND created <= $2 ORDER BY created DESC LIMIT $3", forum, since, limit)
			} else {
				rows, err = db.Query("SELECT * FROM threads WHERE forum = $1 AND created >= $2 ORDER BY created ASC LIMIT $3", forum, since, limit)
			}
		} else {
			if desc == "true" {
				rows, err = db.Query("SELECT * FROM threads WHERE forum = $1 ORDER BY created DESC LIMIT $2", forum, limit)
			} else {
				rows, err = db.Query("SELECT * FROM threads WHERE forum = $1 ORDER BY created ASC LIMIT $2", forum, limit)
			}
		}
	} else {
		if since != "" {
			if desc == "true" {
				rows, err = db.Query("SELECT * FROM threads WHERE forum = $1 AND created <= $2 ORDER BY created DESC", forum, since)
			} else {
				rows, err = db.Query("SELECT * FROM threads WHERE forum = $1 AND created <= $2 ORDER BY created ASC", forum, since)
			}
		} else {
			if desc == "true" {
				rows, err = db.Query("SELECT * FROM threads WHERE forum = $1 ORDER BY created DESC", forum)
			} else {
				rows, err = db.Query("SELECT * FROM threads WHERE forum = $1 ORDER BY created ASC", forum)
			}
		}
	}

	if err != nil {
		return []models.Thread{}
	}
	var thread models.Thread
	var result = make([]models.Thread, 0)
	for rows.Next() {
		err = rows.Scan(&thread.ID, &thread.Slug, &thread.Created, &thread.Message, &thread.Title, &thread.Author, &thread.Forum, &thread.Votes)
		if err != nil {
			return []models.Thread{}
		}
		thread.ID = GetIdByNickname(thread.Author)
		result = append(result, thread)
	}
	return result
}

func GetThreadsByForum(forum string) ([]models.Thread) {
	db := common.GetDB()
	rows, err := db.Query(`SELECT * FROM threads WHERE forum = $1`, forum)
	if err != nil {
		return []models.Thread{}
	}
	var thread models.Thread
	var result = make([]models.Thread, 0)
	for rows.Next() {
		err = rows.Scan(&thread.ID, &thread.Slug, &thread.Created, &thread.Message, &thread.Title, &thread.Author, &thread.Forum, &thread.Votes)
		if err != nil {
			return []models.Thread{}
		}
		thread.ID = GetIdByNickname(thread.Author)
		result = append(result, thread)
	}
	return result
}

func GetSlugCase(slug string) string {
	db := common.GetDB()
	rows, err := db.Query(`SELECT slug FROM forums WHERE slug = $1`, slug)
	if err != nil {
		return  slug
	}
	var result string
	for rows.Next() {
		err = rows.Scan(&result)
	}
	if err != nil {
		return  slug
	}
	return result
}

func SlugExists(slug string) bool {
	db := common.GetDB()
	rows, err := db.Query(`SELECT title FROM forums WHERE slug = $1`, slug)
	if err != nil {
		return false
	}
	var title string
	for rows.Next() {
		err = rows.Scan(&title)
	}
	if err != nil || title == "" {
		return false
	}
	return true
}
