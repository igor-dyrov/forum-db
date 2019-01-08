package getters

import (
	"fmt"
	"strconv"

	"github.com/jackc/pgx"

	"github.com/igor-dyrov/forum-db/src/common"
	"github.com/igor-dyrov/forum-db/src/models"
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

func getAllParents(threadId int, limit, since string, desc bool) string {

	sinceStr := ""
	if since != "" {
		sinceStr = "AND path[1] "
		if desc {
			sinceStr += "< "
		} else {
			sinceStr += "> "
		}
		sinceStr += "(SELECT p.path[1] FROM posts p WHERE p.id = " + since + " )"
	}

	order := "ASC"
	if desc {
		order = "DESC"
	}

	limitStr := ""
	if limit != "" {
		limitStr = "LIMIT " + limit
	}

	return fmt.Sprintf(
		"SELECT id FROM posts WHERE thread = %s AND parent = 0 %s ORDER BY id %s %s",
		strconv.Itoa(threadId),
		sinceStr,
		order,
		limitStr,
	)
}

func GetPostsParentTreeSort(threadID int, limit, since string, desc bool) []models.Post {

	parentsQuery := getAllParents(threadID, limit, since, desc)

	order := "path"
	if desc {
		order = "path[1] DESC, path"
	}

	query := fmt.Sprintf(
		"SELECT created, id, message, parent, author, forum, thread FROM posts WHERE path[1] IN (%s) AND thread = %s ORDER BY %s;",
		parentsQuery,
		strconv.Itoa(threadID),
		order,
	)

	rows, err := common.GetPool().Query(query)
	defer rows.Close()
	PanicIfError(err)

	posts := make([]models.Post, 0)

	for rows.Next() {
		var post models.Post
		PanicIfError(rows.Scan(&post.Created, &post.Id, &post.Message, &post.Parent, &post.Author, &post.Forum, &post.Thread))
		posts = append(posts, post)
	}

	return posts
}
