package getters

import (
	"github.com/igor-dyrov/forum-db/src/common"
	"github.com/igor-dyrov/forum-db/src/models"
)

func GetForumBySlug(slug string) models.Forum {

	rows, err := common.GetPool().Query("SELECT id, slug, title, author, threads, posts from forums WHERE slug = $1", slug)
	defer rows.Close()
	PanicIfError(err)

	var forum models.Forum

	for rows.Next() {
		PanicIfError(rows.Scan(&forum.ID, &forum.Slug, &forum.Title, &forum.Author, &forum.Threads, &forum.Posts))
		return forum
	}
	return forum
}

func GetForumSlug(slug string) string {

	rows, err := common.GetPool().Query(`SELECT slug FROM forums WHERE slug = $1`, slug)
	defer rows.Close()
	PanicIfError(err)

	var result string
	for rows.Next() {
		PanicIfError(rows.Scan(&result))
		return result
	}

	return ""
}

func ForumSlugExists(slug string) bool {

	rows, err := common.GetPool().Query("SELECT slug FROM forums WHERE slug = $1;", slug)
	defer rows.Close()
	PanicIfError(err)

	if rows.Next() {
		var slug string
		PanicIfError(rows.Scan(&slug))
		return true
	}

	return false
}
