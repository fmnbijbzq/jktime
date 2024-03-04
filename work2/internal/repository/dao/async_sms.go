package dao

import (
	"context"
	"time"

	"github.com/ecodeclub/ekit/sqlx"
	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var ErrWaitingSMSNotFound = gorm.ErrRecordNotFound

type AsyncSmsDao interface {
	Insert(ctx context.Context, s AsyncSms) error
	GetWaitingSMS(ctx context.Context) (AsyncSms, error)
	MarkSuccess(ctx context.Context, id int64) error
	MarkFailed(ctx context.Context, id int64) error
}

const (
	// 因为本身状态没有暴露出去，所以不需要在 domain 里面定义
	asyncStatusWaiting = iota
	// 失败了，并且超过了重试次数
	asyncStatusFailed
	asyncStatusSuccess
)

type GORMSmsDao struct {
	db *gorm.DB
}

func NewSmsDao(db *gorm.DB) AsyncSmsDao {
	return &GORMSmsDao{
		db: db,
	}
}

func (dao *GORMSmsDao) GetWaitingSMS(ctx context.Context) (AsyncSms, error) {
	var s AsyncSms
	err := dao.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 为了避免一些偶发性的失败，只寻找1分钟之前的异步短信发送
		now := time.Now().UnixMilli()
		endTime := now - time.Minute.Milliseconds()
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("utime < ? and status = ?", endTime, asyncStatusWaiting).First(&s).Error
		if err != nil {
			return err
		}

		err = tx.Model(&AsyncSms{}).
			Where("id = ?", s.Id).
			Updates(map[string]any{
				"retry_cnt": gorm.Expr("retry_cnt+1"),
				"utime":     now,
			}).Error
		return err
	})

	return s, err
}

func (dao *GORMSmsDao) MarkSuccess(ctx context.Context, id int64) error {
	now := time.Now().UnixMilli()
	return dao.db.WithContext(ctx).Model(&AsyncSms{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"utime":  now,
			"status": asyncStatusSuccess,
		}).Error

}
func (dao *GORMSmsDao) MarkFailed(ctx context.Context, id int64) error {
	now := time.Now().UnixMilli()
	return dao.db.WithContext(ctx).Model(&AsyncSms{}).
		Where("id = ? and `retry_cnt` >= `retry_max`", id).
		Updates(map[string]any{
			"utime":  now,
			"status": asyncStatusFailed,
		}).Error
}

func (dao *GORMSmsDao) Insert(ctx context.Context, s AsyncSms) error {
	err := dao.db.WithContext(ctx).Create(&s).Error
	if me, ok := err.(*mysql.MySQLError); ok {
		const duplicateErr uint16 = 1062
		if me.Number == duplicateErr {
			return ErrDuplicateUser
		}
	}
	return err
}

type AsyncSms struct {
	Id int64 `gorm:"primaryKey;autoIncrement"`
	// 使用我在 ekit 里面支持的 JSON 字段
	Config sqlx.JsonColumn[SmsConfig]
	// 重试次数
	RetryCnt int
	// 重试的最大次数
	RetryMax int
	Status   uint8
	Ctime    int64
	Utime    int64 `gorm:"index"`
}

type SmsConfig struct {
	TplId   string
	Args    []string
	Numbers []string
}
