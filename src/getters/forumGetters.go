package getters

import (
	"github.com/igor-dyrov/forum-db/src/common"
	"github.com/igor-dyrov/forum-db/src/models"
	"github.com/jackc/pgx"
)

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

func GetForumSlug(slug string, conn *pgx.Conn) string {

	rows, err := conn.Query(`SELECT slug FROM forums WHERE slug = $1`, slug)
	defer rows.Close()
	PanicIfError(err)

	var result string
	for rows.Next() {
		err = rows.Scan(&result)
		PanicIfError(err)
		return result
	}
	return ""
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
