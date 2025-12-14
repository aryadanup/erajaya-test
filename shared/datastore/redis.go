package datastore

import (
	"context"
	"strconv"

	"fmt"

	"github.com/redis/go-redis/v9"
)

type RedisDB struct {
	cfg    Config
	client *redis.Client
}

func (r *RedisDB) Connect(ctx context.Context) error {

	var db int
	if r.cfg.DBName == "" {
		db = 0
	} else {
		db, _ = strconv.Atoi(r.cfg.DBName)
	}

	r.client = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", r.cfg.Host, r.cfg.Port),
		Password: r.cfg.Password,
		DB:       db,
	})
	return r.Ping(ctx)
}

func (r *RedisDB) Ping(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}

func (r *RedisDB) Close(ctx context.Context) error {
	return r.client.Close()
}

func (r *RedisDB) GetClient() interface{} {
	return r.client
}
