package domain

import "time"

type User struct {
	Id       int64
	Email    string
	Phone    string
	Password string
	//昵称
	Nickname string
	//生日
	Birthday time.Time
	//个人简介
	AboutMe string
	// UTC 0 的时区
	Ctime time.Time
}
