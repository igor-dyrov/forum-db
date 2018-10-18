package models

type Forum struct {
	Posts int `json:"posts"`
	Slug string `json:"slug"`
	Threads int `json:"threads"`
	Title string `json:"title"`
	Author string `json:"user"`
	ID int `json:"-"`
}
