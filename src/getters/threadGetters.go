package getters

import (
	"strconv"

	"github.com/jackc/pgx"

	"github.com/igor-dyrov/forum-db/src/common"
	"github.com/igor-dyrov/forum-db/src/models"
)

func GetThreadById(id int) (bool, models.Thread) {

	rows, err := common.GetPool().Query("SELECT id, coalesce(slug::text, ''), created, message, title, author, forum, votes FROM threads WHERE id = $1;", id)
	defer rows.Close()
	PanicIfError(err)

	var thread models.Thread

	for rows.Next() {
		PanicIfError(rows.Scan(&thread.ID, &thread.Slug, &thread.Created, &thread.Message, &thread.Title, &thread.Author, &thread.Forum, &thread.Votes))
		return true, thread
	}

	return false, thread
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

func ConnGetThreadBySlug(slug string) *models.Thread {

	rows, err := common.GetPool().Query("SELECT id, slug, created, message, title, author, forum, votes FROM threads WHERE slug = $1", slug)
	defer rows.Close()
	PanicIfError(err)

	result := new(models.Thread)
	for rows.Next() {
		PanicIfError(rows.Scan(&result.ID, &result.Slug, &result.Created, &result.Message, &result.Title, &result.Author, &result.Forum, &result.Votes))
		return result
	}

	return nil
}

func GetThreadBySlugOrID(slugOrId string) *models.Thread {

	thread := new(models.Thread)

	var rows *pgx.Rows
	id, err := strconv.Atoi(slugOrId)
	if err == nil {
		rows, err = common.GetPool().Query("SELECT id, coalesce(slug::text, ''), created, message, title, author, forum, votes FROM threads WHERE id = $1;", id)
	} else {
		rows, err = common.GetPool().Query("SELECT id, coalesce(slug::text, ''), created, message, title, author, forum, votes FROM threads WHERE slug = $1;", slugOrId)
	}
	defer rows.Close()
	PanicIfError(err)

	if rows.Next() {
		PanicIfError(rows.Scan(&thread.ID, &thread.Slug, &thread.Created, &thread.Message, &thread.Title, &thread.Author, &thread.Forum, &thread.Votes))
		return thread
	}

	return nil
}

func GetThreadIDAndForumBySlugOrID(slugOrId string) (bool, int, string) {

	var rows *pgx.Rows
	id, err := strconv.Atoi(slugOrId)
	if err == nil {
		rows, err = common.GetPool().Query("SELECT id, forum FROM threads WHERE id = $1;", id)
	} else {
		rows, err = common.GetPool().Query("SELECT id, forum FROM threads WHERE slug = $1;", slugOrId)
	}
	defer rows.Close()
	PanicIfError(err)

	if rows.Next() {
		var forumSlug string
		var id int
		PanicIfError(rows.Scan(&id, &forumSlug))
		return true, id, forumSlug
	}

	return false, 0, ""
}

func GetThreads(forum string, limit string, since string, desc string) []models.Thread {

	db := common.GetPool()

	var rows *pgx.Rows
	var err error

	if limit != "" {
		if since != "" {
			if desc == "true" {
				rows, err = db.Query("SELECT id, coalesce(slug::text, ''), created, message, title, author, forum, votes FROM threads WHERE forum = $1 AND created <= $2 ORDER BY created DESC LIMIT $3", forum, since, limit)
			} else {
				rows, err = db.Query("SELECT id, coalesce(slug::text, ''), created, message, title, author, forum, votes FROM threads WHERE forum = $1 AND created >= $2 ORDER BY created ASC LIMIT $3", forum, since, limit)
			}
		} else {
			if desc == "true" {
				rows, err = db.Query("SELECT id, coalesce(slug::text, ''), created, message, title, author, forum, votes FROM threads WHERE forum = $1 ORDER BY created DESC LIMIT $2", forum, limit)
			} else {
				rows, err = db.Query("SELECT id, coalesce(slug::text, ''), created, message, title, author, forum, votes FROM threads WHERE forum = $1 ORDER BY created ASC LIMIT $2", forum, limit)
			}
		}
	} else {
		if since != "" {
			if desc == "true" {
				rows, err = db.Query("SELECT id, coalesce(slug::text, ''), created, message, title, author, forum, votes FROM threads WHERE forum = $1 AND created <= $2 ORDER BY created DESC", forum, since)
			} else {
				rows, err = db.Query("SELECT id, coalesce(slug::text, ''), created, message, title, author, forum, votes FROM threads WHERE forum = $1 AND created <= $2 ORDER BY created ASC", forum, since)
			}
		} else {
			if desc == "true" {
				rows, err = db.Query("SELECT id, coalesce(slug::text, ''), created, message, title, author, forum, votes FROM threads WHERE forum = $1 ORDER BY created DESC", forum)
			} else {
				rows, err = db.Query("SELECT id, coalesce(slug::text, ''), created, message, title, author, forum, votes FROM threads WHERE forum = $1 ORDER BY created ASC", forum)
			}
		}
	}
	defer rows.Close()
	PanicIfError(err)

	var thread models.Thread

	var result = make([]models.Thread, 0)

	for rows.Next() {
		PanicIfError(rows.Scan(&thread.ID, &thread.Slug,
			&thread.Created, &thread.Message, &thread.Title, &thread.Author, &thread.Forum, &thread.Votes))
		result = append(result, thread)
	}
	return result
}

func GetThreadsByForum(forum string) []models.Thread {

	rows, err := common.GetPool().Query(
		"SELECT id, coalesce(slug::text, ''), created, message, title, author, forum, votes FROM threads WHERE forum = $1",
		forum,
	)
	defer rows.Close()
	PanicIfError(err)

	var thread models.Thread

	var result = make([]models.Thread, 0)

	for rows.Next() {
		PanicIfError(rows.Scan(&thread.ID, &thread.Slug,
			&thread.Created, &thread.Message, &thread.Title, &thread.Author, &thread.Forum, &thread.Votes))
		result = append(result, thread)
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
