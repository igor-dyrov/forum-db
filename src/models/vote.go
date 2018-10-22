package models

type Vote struct {
	ID int `json:"id"`
	Nickname string `json:"nickname"`
	Voice int `json:"voice"`
	Thread int `json:"thread"`
}