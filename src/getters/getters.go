package getters

import (
	_ "github.com/lib/pq"
	"../models"
	"../common"
)

func GetUserByNick(user models.User) (bool, []models.User) {
	db := common.GetDB()

	rows, err := db.Query(`SELECT * from users WHERE email = $1 OR nickname = $2`, user.Email, user.Nickname)
	result := make([]models.User, 0)

	if err != nil {
		if err != nil {
			return false, result
		}
	}

	for rows.Next() {
		var gotUser models.User
		err = rows.Scan(&gotUser.About, &gotUser.Email, &gotUser.Fullname, &gotUser.Nickname, &gotUser.ID)
		result = append(result, gotUser)
	}

	if err == nil && len(result) != 0 {
		return true, result
	}
	return false, result
}
