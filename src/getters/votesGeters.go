package getters

import (
	"github.com/igor-dyrov/forum-db/src/common"
	"github.com/igor-dyrov/forum-db/src/models"
)

func GetVote(nick string, thread int) models.Vote {
	db := common.GetDB()

	rows, err := db.Query("SELECT id, nickname, voice, thread FROM votes WHERE nickname = $1 AND thread = $2;", nick, thread)
	if err != nil {
		panic(err)
	}

	var result models.Vote
	result.ID = -1

	for rows.Next() {
		err = rows.Scan(&result.ID, &result.Nickname, &result.Voice, &result.Thread)
		if err != nil {
			panic(err)
		}
	}
	return result
}
