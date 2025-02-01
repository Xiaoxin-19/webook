package limiter

import "context"

//go:generate mockgen -source=types.go -package=limitmocks -destination=./mock/limit.mock.go
type Limiter interface {
	Limit(ctx context.Context, key string) (bool, error)
}
