package domain

import "time"

type User struct {
	Id        int64 `json:"-"`
	Email     string
	Phone     string
	Password  string `json:"-"`
	NickName  string
	Birthday  time.Time
	Biography string
	CreatedAt int64 `json:"-"`
	UpdatedAT int64 `json:"-"`
}
