package mocks

import (
	"context"
	"time"

	"github.com/stretchr/testify/mock"
)

type RedisRepository struct {
	mock.Mock
}

func (m *RedisRepository) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	ret := m.Called(ctx, key, value, expiration)
	return ret.Error(0)
}

func (m *RedisRepository) Get(ctx context.Context, key string) (string, error) {
	ret := m.Called(ctx, key)
	return ret.String(0), ret.Error(1)
}

func (m *RedisRepository) Delete(ctx context.Context, key string) error {
	ret := m.Called(ctx, key)
	return ret.Error(0)
}

func (m *RedisRepository) DeleteByPattern(ctx context.Context, pattern string) error {
	ret := m.Called(ctx, pattern)
	return ret.Error(0)
}
