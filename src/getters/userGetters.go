package getters

import (
	"fmt"

	"github.com/igor-dyrov/forum-db/src/common"
	"github.com/igor-dyrov/forum-db/src/models"
	"github.com/jackc/pgx"
)

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

func GetUserByNickname(nickname string) (bool, models.User) {

	rows, err := common.GetPool().Query("SELECT about, email, fullname, nickname, id from users WHERE nickname = $1;", nickname)
	defer rows.Close()
	PanicIfError(err)

	var user models.User

	if rows.Next() {
		PanicIfError(rows.Scan(&user.About, &user.Email, &user.Fullname, &user.Nickname, &user.ID))
		return true, user
	}

	return false, user
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

func joinStrings(nicknames map[string]bool) (s string) {
	sep := ""
	for str, _ := range nicknames {
		s += sep
		sep = ", "
		s += fmt.Sprintf("'%s'", str)
	}
	return
}

func GetUsersByNicknames(nicknames map[string]bool) []string {

	rows, err := common.GetPool().Query("SELECT nickname FROM users WHERE nickname = ANY (ARRAY[" + joinStrings(nicknames) + "])")
	defer rows.Close()
	if err != nil {
		panic(err)
	}

	nicksArray := make([]string, 0, len(nicknames))

	for rows.Next() {
		var nick string
		err := rows.Scan(&nick)
		if err != nil {
			panic(err)
		}
		nicksArray = append(nicksArray, nick)
	}
	return nicksArray
}

func CheckUserByNickname(nickname string, conn *pgx.Conn) bool {

	rows, err := conn.Query("SELECT nickname FROM users WHERE nickname = $1", nickname)
	defer rows.Close()
	PanicIfError(err)

	// var nick string
	if rows.Next() {
		// err := rows.Scan(&nick)
		// PanicIfError(err)
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
