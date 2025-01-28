package dao

import (
	"context"
	"database/sql"
	"errors"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
	"log"
	"time"
)

var (
	ErrDuplicateEmail = errors.New("邮箱冲突")
	ErrRecordNotFound = gorm.ErrRecordNotFound
)

//go:generate mockgen -source=user.go -package=daomocks -destination=./mock/user.mock.go
type UserDAO interface {
	FindByEmail(ctx context.Context, email string) (*User, error)
	Insert(ctx context.Context, user *User) error
	UpdateById(ctx context.Context, u *User) error
	FindById(ctx context.Context, id int64) (*User, error)
	FindByPhone(ctx context.Context, phone string) (*User, error)
}

type GORMUserDAO struct {
	db *gorm.DB
}

func NewGormUserDAO(db *gorm.DB) UserDAO {
	return &GORMUserDAO{db: db}
}

type User struct {
	Id       int64          `gorm:"primaryKey,autoIncrement"`
	Email    sql.NullString `gorm:"unique"`
	Phone    sql.NullString `gorm:"unique"`
	Password string
	Nickname string `gorm:"type=varchar(34)"`
	Birthday int64
	AboutMe  string `gorm:"type=varchar(1024)"`
	// 创建时间, 时区，UTC 0 的毫秒数
	Ctime int64
	// 更新时间
	Utime int64
}

func (dao *GORMUserDAO) FindByEmail(ctx context.Context, email string) (*User, error) {
	u := new(User)
	err := dao.db.WithContext(ctx).Where("email=?", email).First(u).Error
	return u, err
}

func (dao *GORMUserDAO) Insert(ctx context.Context, user *User) error {
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

func (dao *GORMUserDAO) UpdateById(ctx context.Context, u *User) error {
	err := dao.db.WithContext(ctx).Model(u).Updates(u).Error
	if err != nil {
		log.Printf("%v", err)
		return err
	}
	return nil
}

func (dao *GORMUserDAO) FindById(ctx context.Context, id int64) (*User, error) {
	u := new(User)
	err := dao.db.WithContext(ctx).Where("id=?", id).First(u).Error
	return u, err
}

func (dao *GORMUserDAO) FindByPhone(ctx context.Context, phone string) (*User, error) {
	var u User
	err := dao.db.WithContext(ctx).Where("phone=?", phone).First(&u).Error
	return &u, err
}
