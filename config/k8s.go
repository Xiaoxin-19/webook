//go:build k8s

package config

var Config = config{
	DB:    DBConfig{DSN: "host=webook-postgres user=postgres password=postgres dbname=webook port=15432 sslmode=disable TimeZone=Asia/Shanghai"},
	Redis: RedisConfig{Addr: "webook-redis:6380"},
}
