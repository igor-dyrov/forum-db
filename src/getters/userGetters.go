package getters

import (
	"fmt"

	"github.com/igor-dyrov/forum-db/src/common"
	"github.com/igor-dyrov/forum-db/src/models"
	"github.com/jackc/pgx"
)

func GetUserByNickOrEmail(nickname string, email string) (bool, []models.User) {

	rows, err := common.GetPool().Query("SELECT about, email, fullname, nickname, id FROM users WHERE email = $1 OR nickname = $2;", email, nickname)
	defer rows.Close()
	PanicIfError(err)

	result := make([]models.User, 0)

	for rows.Next() {
		var user models.User
		PanicIfError(rows.Scan(&user.About, &user.Email, &user.Fullname, &user.Nickname, &user.ID))
		result = append(result, user)
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

	rows, err := common.GetPool().Query("SELECT nickname from users WHERE email = $1", email)
	defer rows.Close()
	PanicIfError(err)

	var nickname string

	if rows.Next() {
		PanicIfError(rows.Scan(&nickname))
		return nickname
	}
	return ""
}

func UserExists(nickname string) bool {

	rows, err := common.GetPool().Query("SELECT id FROM users WHERE nickname = $1", nickname)
	defer rows.Close()
	PanicIfError(err)

	var id int64
	if rows.Next() {
		PanicIfError(rows.Scan(&id))
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
	PanicIfError(err)

	nicksArray := make([]string, 0, len(nicknames))

	for rows.Next() {
		var nick string
		PanicIfError(rows.Scan(&nick))
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
