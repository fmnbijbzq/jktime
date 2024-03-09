package web

import (
	"time"
)

type ArticleVo struct {
	Id         int64     `json:"id,omitempty"`
	Title      string    `json:"title,omitempty"`
	Content    string    `json:"content,omitempty"`
	Abstact    string    `json:"abstact,omitempty"`
	AuthorId   int64     `json:"author_id,omitempty"`
	AuthorName string    `json:"author_name,omitempty"`
	Status     uint8     `json:"status,omitempty"`
	Utime      time.Time `json:"utime,omitempty"`
	Ctime      time.Time `json:"ctime,omitempty"`
}
