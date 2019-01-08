package getters

import (
	"fmt"
	"strconv"

	"github.com/igor-dyrov/forum-db/src/common"
	"github.com/igor-dyrov/forum-db/src/models"
	"github.com/jackc/pgx"
)

func GetPostByID(id int32) (bool, models.Post) {

	rows, err := common.GetPool().
		Query("select id, author, created, forum, isEdited, message, parent, thread from posts where id = $1;", id)
	defer rows.Close()
	PanicIfError(err)

	if rows.Next() {
		var result models.Post
		err = rows.Scan(&result.Id, &result.Author, &result.Created, &result.Forum,
			&result.IsEdited, &result.Message, &result.Parent, &result.Thread)
		PanicIfError(err)
		return true, result
	}

	return false, models.Post{}
}

func joinInt32(idArray map[int32]bool) (result string) {
	sep := ""
	for id, _ := range idArray {
		result += sep
		sep = ", "
		result += fmt.Sprintf("%d", id)
	}
	return result
}

func GetPostsIDByIDs(idArray map[int32]bool, threadID int) []int32 {

	rows, err := common.GetPool().Query("SELECT id FROM posts WHERE thread = $1 AND id = ANY (ARRAY["+joinInt32(idArray)+"])", threadID)
	defer rows.Close()
	if err != nil {
		panic(err)
	}

	resultArray := make([]int32, 0, len(idArray))

	for rows.Next() {
		var id int32
		err := rows.Scan(&id)
		if err != nil {
			panic(err)
		}
		resultArray = append(resultArray, id)
	}

	return resultArray
}

func GetParentPosts(id int) []models.Post {
	db := common.GetDB()
	var posts []models.Post

	rows, err := db.Query(`SELECT * FROM posts WHERE parent = 0 AND thread = $1 ORDER BY id`, id)
	defer rows.Close()
	PanicIfError(err)

	for rows.Next() {
		var result models.Post
		var gotPath string
		PanicIfError(rows.Scan(&result.Id, &result.Author, &result.Created, &result.Forum,
			&result.IsEdited, &result.Message, &result.Parent, &result.Thread, &gotPath))

		posts = append(posts, result)
	}
	return posts
}

// ---------------------------------------------------------------------------------------------------------

func handlePostRows(rows *pgx.Rows) []models.Post {
	posts := make([]models.Post, 0)

	for rows.Next() {
		var result models.Post
		PanicIfError(rows.Scan(&result.Id, &result.Author, &result.Created, &result.Forum,
			&result.IsEdited, &result.Message, &result.Parent, &result.Thread))
		posts = append(posts, result)
	}

	return posts
}

func GetPostsFlatSort(threadID int, limit, since string, desc bool) []models.Post {

	req := "SELECT id, author, created, forum, isEdited, message, parent, thread FROM posts WHERE thread = $1"

	if limit != "" {
		if !desc {
			if since != "" {
				req += "AND id >" + since + " ORDER BY id ASC LIMIT " + limit
			} else {
				req += "ORDER BY id LIMIT " + limit
			}
		} else {
			if since != "" {
				req += "AND id <" + since + " ORDER BY id DESC LIMIT " + limit
			} else {
				req += "ORDER BY id DESC LIMIT " + limit
			}
		}
	} else {
		if !desc {
			if since != "" {
				req += "AND id >" + since + " ORDER BY id ASC"
			} else {
				req += "ORDER BY id"
			}
		} else {
			if since != "" {
				req += "AND id <" + since + " ORDER BY id DESC"
			} else {
				req += "ORDER BY id DESC"
			}
		}
	}

	rows, err := common.GetPool().Query(req, threadID)
	defer rows.Close()
	PanicIfError(err)

	return handlePostRows(rows)
}

func GetPostsTreeSort(threadID int, limit, since string, desc bool) []models.Post {

	db := common.GetPool()

	var rows *pgx.Rows
	var err error
	var req = "SELECT p1.id, p1.author, p1.created, p1.forum, p1.isEdited, p1.message, p1.parent, p1.thread FROM posts AS p1"

	if limit != "" {
		if !desc {
			if since != "" {
				rows, err = db.Query(req+" JOIN posts AS p2 ON p1.path > p2.path AND p2.id = $2 WHERE p1.thread = $1 ORDER BY p1.path ASC LIMIT $3", threadID, since, limit)
			} else {
				rows, err = db.Query(req+" WHERE thread = $1 ORDER BY p1.path LIMIT $2", threadID, limit)
			}
		} else {
			if since != "" {
				rows, err = db.Query(req+" JOIN posts AS p2 ON p1.path < p2.path AND p2.id = $2 WHERE p1.thread = $1 ORDER BY p1.path DESC LIMIT $3", threadID, since, limit)
			} else {
				rows, err = db.Query(req+" WHERE thread = $1 ORDER BY p1.path DESC LIMIT $2", threadID, limit)
			}
		}
	} else {
		if !desc {
			if since != "" {
				rows, err = db.Query(req+" JOIN posts AS p2 ON p1.path > p2.path AND p2.id = $2 WHERE p1.thread = $1 ORDER BY p1.path ASC", threadID, since)
			} else {
				rows, err = db.Query(req+" WHERE thread = $1 ORDER BY p1.path ASC", threadID)
			}
		} else {
			if since != "" {
				rows, err = db.Query(req+" JOIN posts AS p2 ON p1.path > p2.path AND p2.id = $2 WHERE p1.thread = $1 ORDER BY p1.path DESC", threadID, since)
			} else {
				rows, err = db.Query(req+" WHERE thread = $1 ORDER BY p1.path DESC", threadID)
			}
		}
	}
	defer rows.Close()
	PanicIfError(err)

	return handlePostRows(rows)
}

func GetPostsParentTreeSort(threadID int, limit, since string, desc bool) []models.Post {

	posts := make([]models.Post, 0)

	parentPosts := GetParentPosts(threadID)
	pageSize, _ := strconv.Atoi(limit)
	db := common.GetPool()

	var req = "SELECT p1.id, p1.author, p1.created, p1.forum, p1.isEdited, p1.message, p1.parent, p1.thread FROM posts AS p1"

	if !desc {
		if since == "" {
			for i := range parentPosts {
				if i < pageSize {
					rows, err := db.Query(req+" WHERE p1.thread = $1 AND p1.path[1] = $2 ORDER BY p1.path ASC, p1.id ASC", threadID, parentPosts[i].Id)
					defer rows.Close()
					PanicIfError(err)
					posts = append(posts, handlePostRows(rows)...)
				}
			}
		} else {
			for i := range parentPosts {
				if i <= pageSize {
					rows, err := db.Query(req+" JOIN posts AS p2 ON p1.thread = $1 AND p1.path[1] > p2.path[1]"+
						" AND p2.id = $2 WHERE p1.path[1] = $3 ORDER BY p1.path ASC, p1.id ASC", threadID, since, parentPosts[i].Id)
					defer rows.Close()
					PanicIfError(err)
					posts = append(posts, handlePostRows(rows)...)
				}
			}
		}
	} else {
		if since == "" {
			for i := range parentPosts {
				if i < pageSize {
					var lastParent = parentPosts[len(parentPosts)-1-i].Id
					rows, err := db.Query(req+" WHERE p1.thread = $1 AND p1.path[1] = $2 ORDER BY p1.path ASC, p1.id ASC", threadID, lastParent)
					defer rows.Close()
					PanicIfError(err)
					posts = append(posts, handlePostRows(rows)...)
				}
			}
		} else {
			for i := range parentPosts {
				if i <= pageSize {
					var lastParent = parentPosts[len(parentPosts)-1-i].Id
					rows, err := db.Query(req+" JOIN posts AS p2 ON p1.thread = $1 AND p1.path[1] < p2.path[1]"+
						" AND p2.id = $2 WHERE p1.path[1] = $3 ORDER BY p1.path ASC, p1.id ASC", threadID, since, lastParent)
					defer rows.Close()
					PanicIfError(err)
					posts = append(posts, handlePostRows(rows)...)
				}
			}
		}
	}

	return posts
}
