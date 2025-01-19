//go:build !k8s

package config

var Config = config{
	DB:    DBConfig{DSN: "host=localhost user=postgres password=postgres dbname=webook port=15432 sslmode=disable TimeZone=Asia/Shanghai"},
	Redis: RedisConfig{Addr: "localhost:6380"},
}
