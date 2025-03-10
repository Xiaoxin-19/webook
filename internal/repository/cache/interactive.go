package cache

import (
	"context"
	_ "embed"
	"fmt"
	"github.com/redis/go-redis/v9"
)

var (
	//go:embed lua/incr_cnt.lua
	luaIncrCnt string
)

const fieldReadCnt = "read_cnt"
const fieldLikeCnt = "like_cnt"
const fieldCollectCnt = "collect_cnt"

//go:generate mockgen -source=interactive.go -package=cachemocks -destination=./mock/interactive.mock.go
type InteractiveCache interface {
	IncrReadCntIfPresent(ctx context.Context, biz string, id int64) error
}

type RedisInteractiveCache struct {
	cmd redis.Cmdable
}

func NewRedisInteractiveCache(cmd redis.Cmdable) InteractiveCache {
	return &RedisInteractiveCache{cmd: cmd}
}

func (r *RedisInteractiveCache) IncrReadCntIfPresent(ctx context.Context, biz string, id int64) error {
	_, res := r.cmd.Eval(ctx, luaIncrCnt, []string{r.key(biz, id)}, fieldReadCnt, "1").Int()
	return res
}

func (r *RedisInteractiveCache) key(biz string, id int64) string {
	return fmt.Sprintf("interactive:%s:%d", biz, id)
}
