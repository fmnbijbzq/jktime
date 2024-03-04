package dao

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
)

var ErrDuplicateUser = errors.New("用户冲突")
var ErrUserNotFound = gorm.ErrRecordNotFound

type User struct {
	ID int64 `gorm:"primaryKey;autoIncrement"`
	// 代表可以为NULL的列
	Email sql.NullString `gorm:"unique"`
	Phone sql.NullString `gorm:"unique"`

	// 1 如果查询要求同时使用openid和unionid, 就要创建联合索引
	// 2 如果查询只用 openid，那么就在openid 创建唯一索引，或者<openid, unionid>联合索引
	// 3 如果查询只用 unionid，那么就在unionid 创建唯一索引，或者<unionid, openid>联合索引
	WechatOpenId  sql.NullString `gorm:"unique"`
	WechatUnionId sql.NullString
	Password      string
	NickName      string
	Birthday      time.Time `gorm:"default:(-)"`
	Biography     string
	CreatedAt     int64
	UpdatedAt     int64
}

type UserDao interface {
	Insert(ctx context.Context, u User) error
	FindByEmail(ctx context.Context, email string) (User, error)
	FindByPhone(ctx context.Context, phone string) (User, error)
	FindByWechat(ctx context.Context, openId string) (User, error)
	FindById(ctx context.Context, id int64) (User, error)
	UpdateById(ctx context.Context, u User) error
}

type GORMUserDao struct {
	db *gorm.DB
}

func NewUserDao(db *gorm.DB) UserDao {
	return &GORMUserDao{
		db: db,
	}
}

func (dao *GORMUserDao) Insert(ctx context.Context, u User) error {
	err := dao.db.WithContext(ctx).Create(&u).Error
	if me, ok := err.(*mysql.MySQLError); ok {
		const duplicateErr uint16 = 1062
		if me.Number == duplicateErr {
			return ErrDuplicateUser
		}
	}
	return err
}

func (dao *GORMUserDao) FindByEmail(ctx context.Context, email string) (User, error) {
	var u User
	err := dao.db.WithContext(ctx).Where("email=?", email).First(&u).Error

	return u, err
}

func (dao *GORMUserDao) FindByPhone(ctx context.Context, phone string) (User, error) {
	var u User
	err := dao.db.WithContext(ctx).Where("phone=?", phone).First(&u).Error

	return u, err
}

func (dao *GORMUserDao) FindByWechat(ctx context.Context, openId string) (User, error) {
	var u User
	err := dao.db.WithContext(ctx).Where("openid=?", openId).First(&u).Error

	return u, err
}

func (dao *GORMUserDao) FindById(ctx context.Context, id int64) (User, error) {
	var u User
	err := dao.db.WithContext(ctx).Where("id=?", id).First(&u).Error
	return u, err
}

func (dao *GORMUserDao) UpdateById(ctx context.Context, u User) error {
	err := dao.db.WithContext(ctx).Model(&u).Where("id=?", u.ID).Updates(u).Error

	return err
}
