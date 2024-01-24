package dao

import (
	"context"
	"errors"
	"time"

	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
)

var ErrDuplicateEmail = errors.New("邮箱冲突")

type User struct {
	ID        int64  `gorm:"primaryKey;autoIncrement"`
	Email     string `gorm:"unique"`
	Password  string
	NickName  string
	Birthday  time.Time
	Biography string
	CreatedAt int64
	UpdatedAt int64
}

type UserDao struct {
	db *gorm.DB
}

func NewUserDao(db *gorm.DB) *UserDao {
	return &UserDao{
		db: db,
	}
}

func InitTables(db *gorm.DB) {
	db.AutoMigrate(&User{})
}

func (dao *UserDao) Insert(ctx context.Context, u User) error {
	err := dao.db.WithContext(ctx).Create(&u).Error
	if me, ok := err.(*mysql.MySQLError); ok {
		const duplicateErr uint16 = 1062
		if me.Number == duplicateErr {
			return ErrDuplicateEmail
		}
	}
	return err
}
func (dao *UserDao) FindByEmail(ctx context.Context, email string) (User, error) {
	var u User
	err := dao.db.WithContext(ctx).Where("email=?", email).First(&u).Error

	return u, err
}
func (dao *UserDao) FindById(ctx context.Context, id int64) (User, error) {
	var u User
	err := dao.db.WithContext(ctx).Where("id=?", id).First(&u).Error
	return u, err
}

func (dao *UserDao) UpdateById(ctx context.Context, u User) error {
	err := dao.db.WithContext(ctx).Model(&u).Where("id=?", u.ID).Updates(u).Error

	return err
}
