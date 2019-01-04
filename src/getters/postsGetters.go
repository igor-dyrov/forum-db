package getters

import (
	"strconv"
	"strings"

	"github.com/igor-dyrov/forum-db/src/common"
	"github.com/igor-dyrov/forum-db/src/models"
)

func CheckParent(parentId int, thread int) bool {
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

func GetPathById(id int) []int {
	db := common.GetDB()
	var result []int
	var gotPath string
	err := db.QueryRow(`SELECT path FROM posts WHERE id = $1`, id).Scan(&gotPath)
	if len(gotPath) > 0 {
		IDs := strings.Split(gotPath[1:len(gotPath)-1], ",")
		for index := range IDs {
			item, _ := strconv.Atoi(IDs[index])
			result = append(result, item)
		}
	}
	if err != nil {
		return []int{}
	}
	return result
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
				item, _ := strconv.Atoi(IDs[index])
				result.Path = append(result.Path, item)
			}
		}
		posts = append(posts, result)
	}
	return posts
}
