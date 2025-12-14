package repository

import (
	"context"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisRepository interface {
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Get(ctx context.Context, key string) (string, error)
	Delete(ctx context.Context, keyOrPattern string) error
}

type redisRepository struct {
	client *redis.Client
}

func NewRedisRepository(client *redis.Client) RedisRepository {
	return &redisRepository{
		client: client,
	}
}

func (r *redisRepository) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return r.client.Set(ctx, key, value, expiration).Err()
}

func (r *redisRepository) Get(ctx context.Context, key string) (string, error) {
	return r.client.Get(ctx, key).Result()
}

func (r *redisRepository) Delete(ctx context.Context, keyOrPattern string) error {
	if strings.Contains(keyOrPattern, "*") {
		return r.deleteByPatternInternal(ctx, keyOrPattern)
	}

	return r.client.Del(ctx, keyOrPattern).Err()
}

func (r *redisRepository) deleteByPatternInternal(ctx context.Context, pattern string) error {
	iter := r.client.Scan(ctx, 0, pattern, 0).Iterator()
	const batchSize = 100
	var keys []string

	for iter.Next(ctx) {
		keys = append(keys, iter.Val())

		if len(keys) >= batchSize {
			if err := r.client.Del(ctx, keys...).Err(); err != nil {
				return err
			}
			keys = keys[:0]
		}
	}

	if err := iter.Err(); err != nil {
		return err
	}

	if len(keys) > 0 {
		if err := r.client.Del(ctx, keys...).Err(); err != nil {
			return err
		}
	}

	return nil
}
