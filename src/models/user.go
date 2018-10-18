package models

type User struct {
	About string `json:"about"`
	Email string `json:"email"`
	Fullname string `json:"fullname"`
	Nickname string `json:"nickname"`
	ID int64 `json:"-"`
}
