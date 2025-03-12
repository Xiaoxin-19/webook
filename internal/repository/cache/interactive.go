package cache

import (
	"context"
	_ "embed"
	"fmt"
	"github.com/redis/go-redis/v9"
	"strconv"
	"time"
	"webok/internal/domain"
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
	IncrLikeCntIfPresent(ctx context.Context, biz string, id int64) error
	DecrLikeCntIfPresent(ctx context.Context, biz string, id int64) error
	IncrCollectionCntIfPresent(ctx context.Context, biz string, id int64) error
	Get(ctx context.Context, biz string, id int64) (domain.Interactive, error)
	Set(ctx context.Context, biz string, id int64, ie domain.Interactive) error
}

type RedisInteractiveCache struct {
	cmd redis.Cmdable
}

func (r *RedisInteractiveCache) Set(ctx context.Context, biz string, id int64, ie domain.Interactive) error {
	_, err := r.cmd.HSet(ctx, r.key(biz, id), fieldLikeCnt, ie.LikeCnt,
		fieldReadCnt, ie.ReadCnt,
		fieldCollectCnt, ie.CollectCnt).Result()
	if err != nil {
		return err
	}
	return r.cmd.Expire(ctx, r.key(biz, id), time.Minute*15).Err()
}

func (r *RedisInteractiveCache) Get(ctx context.Context, biz string, id int64) (domain.Interactive, error) {
	res, err := r.cmd.HGetAll(ctx, r.key(biz, id)).Result()
	if err != nil {
		return domain.Interactive{}, err
	}

	if len(res) == 0 {
		return domain.Interactive{}, ErrKeyNotExist
	}

	var interactive domain.Interactive
	interactive.LikeCnt, _ = strconv.ParseInt(res[fieldLikeCnt], 10, 64)
	interactive.ReadCnt, _ = strconv.ParseInt(res[fieldReadCnt], 10, 64)
	interactive.CollectCnt, _ = strconv.ParseInt(res[fieldCollectCnt], 10, 64)

	return interactive, nil
}

func (r *RedisInteractiveCache) IncrCollectionCntIfPresent(ctx context.Context, biz string, id int64) error {
	_, res := r.cmd.Eval(ctx, luaIncrCnt, []string{r.key(biz, id)}, fieldCollectCnt, 1).Int()
	return res
}

func (r *RedisInteractiveCache) IncrLikeCntIfPresent(ctx context.Context, biz string, id int64) error {
	_, res := r.cmd.Eval(ctx, luaIncrCnt, []string{r.key(biz, id)}, fieldLikeCnt, 1).Int()
	return res
}

func (r *RedisInteractiveCache) DecrLikeCntIfPresent(ctx context.Context, biz string, id int64) error {
	_, res := r.cmd.Eval(ctx, luaIncrCnt, []string{r.key(biz, id)}, fieldLikeCnt, -1).Int()
	return res
}

func NewRedisInteractiveCache(cmd redis.Cmdable) InteractiveCache {
	return &RedisInteractiveCache{cmd: cmd}
}

func (r *RedisInteractiveCache) IncrReadCntIfPresent(ctx context.Context, biz string, id int64) error {
	_, res := r.cmd.Eval(ctx, luaIncrCnt, []string{r.key(biz, id)}, fieldReadCnt, 1).Int()
	return res
}

func (r *RedisInteractiveCache) key(biz string, id int64) string {
	return fmt.Sprintf("interactive:%s:%d", biz, id)
}
