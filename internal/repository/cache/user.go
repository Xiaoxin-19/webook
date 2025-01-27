package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/redis/go-redis/v9"
	"time"
	"webok/internal/domain"
)

var ErrKeyNotExist = redis.Nil

type UserCache struct {
	cmd        redis.Cmdable // 为什么使用接口：1.面向接口编程，eg：如何要兼容集群怎么办？
	expiration time.Duration
}

func NewUserCache(cmd redis.Cmdable) *UserCache {
	//为什么不传入一个地址，在此函数内初始化redis？
	//松耦合，一定不要自己初始化，依赖要从外面进来，
	return &UserCache{
		cmd:        cmd,
		expiration: time.Minute * 15,
	}
}

func (c *UserCache) Get(ctx context.Context, uid int64) (*domain.User, error) {
	key := c.key(uid)
	data, err := c.cmd.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}
	var u domain.User
	err = json.Unmarshal([]byte(data), &u)
	return &u, err
}

func (c *UserCache) Set(ctx context.Context, du *domain.User) error {
	key := c.key(du.Id)
	data, err := json.Marshal(*du)
	if err != nil {
		return err
	}
	return c.cmd.Set(ctx, key, data, c.expiration).Err()
}

func (c *UserCache) key(uid int64) string {
	return fmt.Sprintf("user:info:%d", uid)
}
