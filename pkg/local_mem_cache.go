package localMemCache

import (
	"errors"
	"log"
	"sync"
	"time"
)

var ErrKeyAlreadyExists = errors.New("key already exists")
var ErrKeyNotFound = errors.New("key not found")

type Item struct {
	Expire time.Time
	Data   any
}
type LocalMemCache struct {
	sync.Mutex
	CleanInterval time.Duration
	cache         map[string]Item
}

func NewLocalMemCache(cleanInterval time.Duration) *LocalMemCache {
	cache := LocalMemCache{
		CleanInterval: cleanInterval,
		cache:         make(map[string]Item),
	}
	go func(c *LocalMemCache) {
		for {
			time.Sleep(10 * time.Second)
			cache.Clean()
		}

	}(&cache)
	return &cache
}

func (c *LocalMemCache) Add(key string, value any, expire time.Duration) error {
	c.Lock()
	defer c.Unlock()
	_, ok := c.cache[key]
	if ok {
		return ErrKeyAlreadyExists
	}
	c.cache[key] = Item{
		Expire: time.Now().Add(expire),
		Data:   value,
	}
	return nil
}

func (c *LocalMemCache) Update(key string, value any) error {
	c.Lock()
	defer c.Unlock()
	res, ok := c.cache[key]
	if !ok {
		return ErrKeyNotFound
	}
	c.cache[key] = Item{
		Expire: res.Expire,
		Data:   value,
	}
	return nil
}

func (c *LocalMemCache) Del(key string) {
	c.Lock()
	defer c.Unlock()
	delete(c.cache, key)
}

func (c *LocalMemCache) Clean() {
	c.Lock()
	defer c.Unlock()
	for k, v := range c.cache {
		if v.Expire.Before(time.Now()) {
			delete(c.cache, k)
		}
		log.Printf("[%v]:%v", k, v.Data)
	}
}

func (c *LocalMemCache) Get(key string) (Item, bool) {
	c.Lock()
	defer c.Unlock()
	item, ok := c.cache[key]
	return item, ok
}
