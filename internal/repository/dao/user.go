package dao

import (
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
	"log"
	"time"
)

var (
	ErrDuplicateEmail = errors.New("邮箱冲突")
	ErrRecordNotFound = gorm.ErrRecordNotFound
)

type UserDAO struct {
	db *gorm.DB
}

func NewUserDAO(db *gorm.DB) *UserDAO {
	return &UserDAO{db: db}
}

type User struct {
	Id       int64  `gorm:"primaryKey,autoIncrement"`
	Email    string `gorm:"unique"`
	Password string
	Nickname string `gorm:"type=varchar(34)"`
	Birthday int64
	Brief    string `gorm:"type=varchar(1024)"`
	// 创建时间, 时区，UTC 0 的毫秒数
	Ctime int64
	// 更新时间
	Utime int64
}

func (dao *UserDAO) FindByEmail(ctx *gin.Context, email string) (*User, error) {
	u := new(User)
	err := dao.db.WithContext(ctx).Where("email=?", email).First(u).Error
	return u, err
}

func (dao *UserDAO) Insert(ctx context.Context, user *User) error {
	now := time.Now().UnixMilli()
	user.Ctime = now
	user.Utime = now
	err := dao.db.WithContext(ctx).Create(user).Error

	var pe *pgconn.PgError
	if errors.As(err, &pe) {
		duplicateErr := "23505"
		if duplicateErr == pe.Code {
			return ErrDuplicateEmail
		}
	}

	return err
}

func (dao *UserDAO) UpdateById(ctx context.Context, u *User) error {
	err := dao.db.WithContext(ctx).Model(u).Updates(u).Error
	if err != nil {
		log.Printf("%v", err)
		return err
	}
	return nil
}

func (dao *UserDAO) FindById(ctx *gin.Context, id int64) (*User, error) {
	u := new(User)
	err := dao.db.WithContext(ctx).Where("id=?", id).First(u).Error
	return u, err
}
