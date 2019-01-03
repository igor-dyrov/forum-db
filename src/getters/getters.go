package getters

import (
	"strconv"
	"strings"

	"database/sql"

	_ "github.com/lib/pq"

	"github.com/igor-dyrov/forum-db/src/common"
	"github.com/igor-dyrov/forum-db/src/models"
)

func PanicIfError(err error) {
	if err != nil {
		panic(err)
	}
}

func GetUserByNickOrEmail(nickname string, email string) (bool, []models.User) {
	db := common.GetDB()

	rows, err := db.Query("SELECT about, email, fullname, nickname, id FROM users WHERE email = $1 OR nickname = $2;", email, nickname)
	PanicIfError(err)

	result := make([]models.User, 0)

	for rows.Next() {
		var gotUser models.User
		err := rows.Scan(&gotUser.About, &gotUser.Email, &gotUser.Fullname, &gotUser.Nickname, &gotUser.ID)
		PanicIfError(err)

		result = append(result, gotUser)
	}

	if len(result) != 0 {
		return true, result
	}
	return false, nil
}

func GetUserByNick(nickname string) models.User {
	db := common.GetDB()

	rows, _ := db.Query(`SELECT * from users WHERE nickname = $1`, nickname)
	var gotUser models.User
	//if err != nil {
	//	return result
	//}

	for rows.Next() {
		rows.Scan(&gotUser.About, &gotUser.Email, &gotUser.Fullname, &gotUser.Nickname, &gotUser.ID)
	}

	return gotUser
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
	rows, err := db.Query("SELECT id FROM users WHERE nickname = $1", nickname)
	defer rows.Close()
	if err != nil {
		panic(err)
	}

	var id int64
	if rows.Next() {
		err := rows.Scan(&id)
		if err != nil {
			panic(err)
		}
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

func GetSlugCase(slug string) string {
	db := common.GetDB()
	rows, err := db.Query(`SELECT slug FROM forums WHERE slug = $1`, slug)
	if err != nil {
		return slug
	}
	var result string
	for rows.Next() {
		err = rows.Scan(&result)
	}
	if err != nil {
		return slug
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

func GetVote(nick string, thread int) models.Vote {
	db := common.GetDB()
	rows, err := db.Query(`SELECT * FROM votes WHERE nickname = $1 AND thread = $2`, nick, thread)
	var result models.Vote
	result.ID = -1
	if err != nil {
		return result
	}
	for rows.Next() {
		err = rows.Scan(&result.ID, &result.Nickname, &result.Voice, &result.Thread)
	}
	return result
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

func GetThreadById(id int) models.Thread {
	db := common.GetDB()
	var result models.Thread
	rows, _ := db.Query(`SELECT * FROM threads WHERE id = $1`, id)

	for rows.Next() {
		rows.Scan(&result.ID, &result.Slug, &result.Created, &result.Message, &result.Title, &result.Author, &result.Forum, &result.Votes)

	}
	return result
}
