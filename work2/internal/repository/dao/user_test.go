package dao_test

import (
	"context"
	"database/sql"
	"errors"
	"example/wb/internal/repository/dao"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/mysql"

	mysqlDriver "github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
)

func TestGORMUserDao_Insert(t *testing.T) {
	testCase := []struct {
		name string

		mock func(t *testing.T) *sql.DB
		ctx  context.Context
		user dao.User

		wantErr error
	}{
		// TODO: Add test cases.
		{
			name: "插入成功",
			mock: func(t *testing.T) *sql.DB {
				db, mock, err := sqlmock.New()
				assert.NoError(t, err)

				mockRes := sqlmock.NewResult(234, 342)
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO .*").
					WillReturnResult(mockRes)
				mock.ExpectCommit()
				return db
			},
			ctx: context.Background(),
			user: dao.User{
				Email: sql.NullString{
					String: "adfsa@daf/com",
					Valid:  true,
				},
				Password: "2143",
			},
		},
		{
			name: "邮箱冲突",
			mock: func(t *testing.T) *sql.DB {
				db, mock, err := sqlmock.New()
				assert.NoError(t, err)
				// 这边要求传入的是 sql 的正则表达式
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO .*").
					WillReturnError(&mysqlDriver.MySQLError{Number: 1062})
				mock.ExpectRollback()
				return db
			},
			ctx: context.Background(),
			user: dao.User{
				NickName: "Tom",
			},
			wantErr: dao.ErrDuplicateUser,
		},
		{
			name: "数据库出错",
			mock: func(t *testing.T) *sql.DB {
				db, mock, err := sqlmock.New()
				assert.NoError(t, err)
				// 这边要求传入的是 sql 的正则表达式
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO .*").
					WillReturnError(errors.New("数据库error"))
				mock.ExpectRollback()
				return db
			},
			ctx: context.Background(),
			user: dao.User{
				NickName: "Tom",
			},
			wantErr: errors.New("数据库error"),
		},
	}
	for _, tt := range testCase {
		t.Run(tt.name, func(t *testing.T) {
			sqlDB := tt.mock(t)
			db, err := gorm.Open(mysql.New(mysql.Config{
				Conn:                      sqlDB,
				SkipInitializeWithVersion: true,
			}), &gorm.Config{
				DisableAutomaticPing:     true,
				DisableNestedTransaction: true,
			})
			assert.NoError(t, err)
			dao := dao.NewUserDao(db)
			err = dao.Insert(tt.ctx, tt.user)
			assert.Equal(t, tt.wantErr, err)
		})
	}
}
