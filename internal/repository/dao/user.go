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
	FindByWechat(ctx context.Context, openId string) (User, error)
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
	// 1 如果查询要求同时使用 openid 和 unionid，就要创建联合唯一索引
	// 2 如果查询只用 openid，那么就在 openid 上创建唯一索引，或者 <openid, unionId> 联合索引
	// 3 如果查询只用 unionid，那么就在 unionid 上创建唯一索引，或者 <unionid, openid> 联合索引
	WechatOpenId  sql.NullString `gorm:"unique"`
	WechatUnionId sql.NullString
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
func (dao *GORMUserDAO) FindByWechat(ctx context.Context, openId string) (User, error) {
	var u User
	err := dao.db.WithContext(ctx).Where("wechat_open_id=?", openId).First(&u).Error
	return u, err
}
