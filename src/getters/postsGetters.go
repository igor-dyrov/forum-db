package getters

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/igor-dyrov/forum-db/src/common"
	"github.com/igor-dyrov/forum-db/src/models"
)

func CheckParent(parentId int32, thread int) bool {
	db := common.GetDB()
	rows, err := db.Query(`SELECT thread FROM posts WHERE id = $1`, parentId)
	if err != nil {
		return false
	}
	var parentThread int
	for rows.Next() {
		rows.Scan(&parentThread)
	}
	if err != nil {
		return false
	}
	return parentThread == thread
}

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
	if err != nil {
		return posts
	}
	for rows.Next() {
		var result models.Post
		var gotPath string
		err = rows.Scan(&result.Id, &result.Author, &result.Created, &result.Forum,
			&result.IsEdited, &result.Message, &result.Parent, &result.Thread, &gotPath)
		if len(gotPath) > 2 {
			IDs := strings.Split(gotPath[1:len(gotPath)-1], ",")
			for index := range IDs {
				item, err := strconv.ParseInt(IDs[index], 10, 32)
				PanicIfError(err)
				result.Path = append(result.Path, int32(item))
			}
		}
		posts = append(posts, result)
	}
	return posts
}
