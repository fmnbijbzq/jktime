package domain

import "time"

type Article struct {
	Id      int64
	Title   string
	Content string
	Author  Author
	Status  ArticleStatus
	Utime   time.Time
	Ctime   time.Time
}

func (a *Article) Abstact() string {
	str := []rune(a.Content)
	// 只取一部分
	if len(str) > 128 {
		str = str[:128]
	}
	return string(str)
}

type ArticleStatus uint8

const (
	// 这是一个位置状态
	ArticleStatusUnknown = iota
	// 未发表
	ArticleStatusUnpublished
	// 已发表
	ArticleStatusPublished
	// 仅自己可见
	ArticleStatusPrivate
)

type Author struct {
	Id   int64
	Name string
}
