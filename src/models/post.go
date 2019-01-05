package models

import "time"

type Post struct {
	Author   string    `json:"author"`
	Created  time.Time `json:"created"`
	Forum    string    `json:"forum"`
	Id       int       `json:"id"`
	Message  string    `json:"message"`
	Thread   int       `json:"thread"`
	Parent   int       `json:"parent"`
	Path     []int     `json:"path"`
	IsEdited bool      `json:"isEdited"`
}

type PostDetails struct {
	Post   *Post   `json:"post"`
	Thread *Thread `json:"thread"`
	Forum  *Forum  `json:"forum"`
	Author *User   `json:"author"`
}
