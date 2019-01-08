package getters

import (
	"github.com/igor-dyrov/forum-db/src/common"
	"github.com/igor-dyrov/forum-db/src/models"
)

func GetVote(nick string, thread int) models.Vote {

	rows, err := common.GetPool().Query("SELECT id, nickname, voice, thread FROM votes WHERE nickname = $1 AND thread = $2;", nick, thread)
	defer rows.Close()
	PanicIfError(err)

	var result models.Vote
	result.ID = -1

	for rows.Next() {
		PanicIfError(rows.Scan(&result.ID, &result.Nickname, &result.Voice, &result.Thread))
		return result
	}

	return result
}
