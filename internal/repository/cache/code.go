package cache

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"github.com/redis/go-redis/v9"
	"log"
	"sync"
	"time"
	localMemCache "webok/pkg"
)

var (
	//go:embed lua/set_code.lua
	luaSetCode string
	//go:embed lua/verify_code.lua
	luaVerifyCode string

	ErrCodeSendTooMany   = errors.New("发送太频繁")
	ErrCodeVerifyTooMany = errors.New("验证太频繁")
)

type CodeCache interface {
	Set(ctx context.Context, biz, phone, code string) error
	Verify(ctx context.Context, biz, phone, code string) (bool, error)
}

type CodeRedisCache struct {
	cmd redis.Cmdable
}

func NewCodeRedisCache(cmd redis.Cmdable) CodeCache {
	return &CodeRedisCache{
		cmd: cmd,
	}
}

// Set 会返回ErrCodeSendTooMany错误
func (c *CodeRedisCache) Set(ctx context.Context, biz, phone, code string) error {
	res, err := c.cmd.Eval(ctx, luaSetCode, []string{key(biz, phone)}, code).Int()
	if err != nil {
		// 调用 redis 出了问题
		return err
	}
	switch res {
	case -2:
		return errors.New("验证码存在，但是没有过期时间")
	case -1:
		return ErrCodeSendTooMany
	default:
		return nil
	}
}

func (c *CodeRedisCache) Verify(ctx context.Context, biz, phone, code string) (bool, error) {
	res, err := c.cmd.Eval(ctx, luaVerifyCode, []string{key(biz, phone)}, code).Int()
	if err != nil {
		// 调用 redis 出了问题
		return false, err
	}
	switch res {
	case -2:
		return false, nil
	case -1:
		return false, ErrCodeVerifyTooMany
	default:
		return true, nil
	}
}

func key(biz, phone string) string {
	return fmt.Sprintf("phone_code:%s:%s", biz, phone)
}

type CodeLocalMemCache struct {
	sync.Mutex
	cache *localMemCache.LocalMemCache
}

func NewCodeLocalMemCache(cache *localMemCache.LocalMemCache) *CodeLocalMemCache {
	return &CodeLocalMemCache{cache: cache}
}

type codeItem struct {
	code string
	// 可验证次数
	cnt int
}

func (c *CodeLocalMemCache) Set(ctx context.Context, biz, phone, code string) error {
	c.Lock()
	defer c.Unlock()
	ckey := key(biz, phone)
	res, ok := c.cache.Get(ckey)
	if !ok {
		err := c.cache.Add(ckey, codeItem{
			code: code,
			cnt:  3,
		}, time.Minute*10)
		if err != nil {
			return err
		}
		return nil
	}

	if res.Expire.Sub(time.Now()) > time.Minute*9 {
		log.Printf("ErrCodeSendTooMany\n")
		return ErrCodeSendTooMany
	}

	//满足发送间隔，发送
	return c.cache.Add(ckey, codeItem{
		code: code,
		cnt:  3,
	}, time.Minute*10)
}

func (c *CodeLocalMemCache) Verify(ctx context.Context, biz, phone, code string) (bool, error) {
	c.Lock()
	defer c.Unlock()

	ckey := key(biz, phone)
	res, ok := c.cache.Get(ckey)
	if !ok || res.Expire.Sub(time.Now()) > time.Minute*10 {
		return false, nil
	}
	item, ok := res.Data.(codeItem)
	if !ok {
		return false, errors.New("系统错误")
	}

	if item.cnt <= 0 {
		return false, ErrCodeVerifyTooMany
	}

	if item.code != code {
		item.cnt--
		err := c.cache.Update(ckey, item)
		if err != nil {
			return false, err
		}
		return false, nil
	}
	c.cache.Del(ckey)
	return true, nil
}
