package getters

import (
	"database/sql"
	"strconv"

	"github.com/jackc/pgx"

	"github.com/igor-dyrov/forum-db/src/common"
	"github.com/igor-dyrov/forum-db/src/models"
)

func GetThreadById(id int) models.Thread {
	db := common.GetDB()
	var result models.Thread
	rows, _ := db.Query(`SELECT * FROM threads WHERE id = $1`, id)

	for rows.Next() {
		rows.Scan(&result.ID, &result.Slug, &result.Created, &result.Message, &result.Title, &result.Author, &result.Forum, &result.Votes)

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

func GetIdBySlug(slug string) int {
	db := common.GetDB()
	rows, err := db.Query(`SELECT id FROM threads WHERE slug = $1`, slug)
	if err != nil {
		return -1
	}
	var result = -1
	for rows.Next() {
		err = rows.Scan(&result)
	}
	if err != nil {
		return -1
	}
	return result
}

func ConnGetThreadBySlug(slug string, conn *pgx.Conn) *models.Thread {

	rows, err := conn.Query("SELECT id, slug, created, message, title, author, forum, votes FROM threads WHERE slug = $1", slug)
	defer rows.Close()
	PanicIfError(err)

	result := new(models.Thread)
	for rows.Next() {
		err = rows.Scan(&result.ID, &result.Slug, &result.Created, &result.Message, &result.Title, &result.Author, &result.Forum, &result.Votes)
		PanicIfError(err)
		return result
	}
	return nil
}

func GetThreadBySlugOrID(slugOrId string) *models.Thread {
	db := common.GetDB()

	thread := new(models.Thread)

	var rows *sql.Rows
	id, err := strconv.Atoi(slugOrId)
	if err == nil {
		rows, err = db.Query("SELECT id, coalesce(slug::text, ''), created, message, title, author, forum, votes FROM threads WHERE id = $1;", id)
	} else {
		rows, err = db.Query("SELECT id, coalesce(slug::text, ''), created, message, title, author, forum, votes FROM threads WHERE slug = $1;", slugOrId)
	}
	defer rows.Close()

	if err != nil {
		panic(err)
	}

	if rows.Next() {
		err = rows.Scan(&thread.ID, &thread.Slug, &thread.Created, &thread.Message, &thread.Title, &thread.Author, &thread.Forum, &thread.Votes)
		if err != nil {
			panic(err)
		}
	} else {
		return nil
	}

	return thread
}

func GetThreads(forum string, limit string, since string, desc string) []models.Thread {
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
		result = append(result, thread)
	}
	return result
}

func GetThreadsByForum(forum string) []models.Thread {
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

func GetSlugById(id int) string {
	db := common.GetDB()
	rows, err := db.Query(`SELECT forum FROM threads WHERE id = $1`, id)
	if err != nil {
		return ""
	}
	var result string
	for rows.Next() {
		err = rows.Scan(&result)
	}
	if err != nil {
		return ""
	}
	return result
}

func GetThreadId(forum string) int {
	db := common.GetDB()
	rows, err := db.Query(`SELECT id FROM threads WHERE slug = $1`, forum)
	if err != nil {
		return -1
	}
	var result = -1
	for rows.Next() {
		err = rows.Scan(&result)
	}
	if err != nil {
		return -1
	}
	return result
}

func GetThreadSlug(slug string) string {
	db := common.GetDB()
	rows, err := db.Query(`SELECT forum FROM threads WHERE slug = $1`, slug)
	if err != nil {
		return ""
	}
	var result string
	for rows.Next() {
		err = rows.Scan(&result)
	}
	if err != nil {
		return ""
	}
	return result
}

func ThreadExists(id int) bool {
	db := common.GetDB()
	rows, err := db.Query(`SELECT id FROM threads WHERE id = $1`, id)
	if err != nil {
		return false
	}
	var Id = -1
	for rows.Next() {
		err = rows.Scan(&Id)
	}
	if err != nil {
		return false
	}
	if Id != -1 {
		return true
	}
	return false
}
