package ioc

import (
	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
)

func InitRedis() redis.Cmdable {
	type config struct {
		Url string `yaml:"Url"`
	}
	var c config
	err := viper.UnmarshalKey("redis", &c)
	if err != nil {
		panic(err)
	}
	return redis.NewClient(&redis.Options{
		Addr: c.Url,
	})
}
