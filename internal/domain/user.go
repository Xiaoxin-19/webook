package domain

import "time"

type User struct {
	Id       int64
	Email    string
	Password string
	//昵称
	Nickname string
	//生日
	Birthday int64
	//个人简介
	Brief string
	// UTC 0 的时区
	Ctime time.Time
}
