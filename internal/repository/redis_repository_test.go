package repository

import (
	"context"
	"erajaya-test/shared/constant"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/go-redis/redismock/v9"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/suite"
)

type RedisSuite struct {
	suite.Suite
	client *redis.Client
	mock   redismock.ClientMock
	repo   RedisRepository
}

func (s *RedisSuite) SetupTest() {
	var err error
	s.client, s.mock = redismock.NewClientMock()
	s.Require().NoError(err)

	s.repo = NewRedisRepository(s.client)
}

func (s *RedisSuite) TearDownTest() {
	s.client.Close()
}

func (s *RedisSuite) TestSetWithKey() {
	ctx := context.Background()
	key := fmt.Sprintf("%s:1:10", constant.RedisKeyProductList)
	val := `{
		"products": [
			{
			"id": 1,
			"name": "LG TV 14 Inci",
			"price": 5000000,
			"description": "LG TV 14 Inci Full HD",
			"quantity": 10
			}
		],
		"total": 9
		}`
	ttl := time.Minute

	s.Run("Success", func() {
		s.mock.ExpectSet(key, val, ttl).SetVal(val)
		err := s.repo.Set(ctx, key, val, ttl)
		s.NoError(err)
	})

	s.Run("Error", func() {
		s.mock.ExpectSet(key, val, ttl).SetErr(errors.New("redis down"))
		err := s.repo.Set(ctx, key, val, ttl)
		s.Error(err)
	})
}

func (s *RedisSuite) TestGetWithKey() {
	ctx := context.Background()
	key := fmt.Sprintf("%s:1:10", constant.RedisKeyProductList)
	val := `{
		"products": [
			{
			"id": 1,
			"name": "LG TV 14 Inci",
			"price": 5000000,
			"description": "LG TV 14 Inci Full HD",
			"quantity": 10
			}
		],
		"total": 9
		}`

	s.Run("Success", func() {
		s.mock.ExpectGet(key).SetVal(val)
		res, err := s.repo.Get(ctx, key)
		s.NoError(err)
		s.Equal(val, res)
	})

	s.Run("Error", func() {
		s.mock.ExpectGet(key).SetErr(errors.New("nil"))
		res, err := s.repo.Get(ctx, key)
		s.Error(err)
		s.Empty(res)
	})
}

func (s *RedisSuite) TestDeleteWithKey() {
	ctx := context.Background()
	key := fmt.Sprintf("%s:1:10", constant.RedisKeyProductList)

	s.Run("Success", func() {
		s.mock.ExpectDel(key).SetVal(1)
		err := s.repo.Delete(ctx, key)
		s.NoError(err)
	})

	s.Run("Error", func() {
		s.mock.ExpectDel(key).SetErr(errors.New("fail"))
		err := s.repo.Delete(ctx, key)
		s.Error(err)
	})
}

func (s *RedisSuite) TestDelete_Pattern() {
	ctx := context.Background()
	pattern := "products:*"

	generateKeys := func(count int) []string {
		keys := make([]string, count)
		for i := 0; i < count; i++ {
			keys[i] = fmt.Sprintf("products:%d", i)
		}
		return keys
	}

	s.Run("Success - Key Found (<100)", func() {
		keys := []string{"products:1", "products:2"}

		s.mock.ExpectScan(0, pattern, 0).SetVal(keys, 0)

		s.mock.ExpectDel(keys...).SetVal(2)

		err := s.repo.Delete(ctx, pattern)
		s.NoError(err)
	})

	s.Run("Success - Key Found (>100)", func() {
		totalKeys := 105
		keys := generateKeys(totalKeys)

		s.mock.ExpectScan(0, pattern, 0).SetVal(keys, 0)

		s.mock.ExpectDel(keys[:100]...).SetVal(100)
		s.mock.ExpectDel(keys[100:]...).SetVal(5)

		err := s.repo.Delete(ctx, pattern)
		s.NoError(err)
	})

	s.Run("Error - Scan Error", func() {
		s.mock.ExpectScan(0, pattern, 0).SetErr(errors.New("scan fail"))

		err := s.repo.Delete(ctx, pattern)
		s.Error(err)
		s.Equal("scan fail", err.Error())
	})

	s.Run("Error - Delete Error (Inside Loop)", func() {
		keys := generateKeys(100)

		s.mock.ExpectScan(0, pattern, 0).SetVal(keys, 0)
		s.mock.ExpectDel(keys...).SetErr(errors.New("del fail"))

		err := s.repo.Delete(ctx, pattern)
		s.Error(err)
		s.Equal("del fail", err.Error())
	})

	s.Run("Error - Delete Error (Remainder)", func() {
		keys := []string{"p:1"}

		s.mock.ExpectScan(0, pattern, 0).SetVal(keys, 0)
		s.mock.ExpectDel(keys...).SetErr(errors.New("del remainder fail"))

		err := s.repo.Delete(ctx, pattern)
		s.Error(err)
	})
}

func TestRedisSuite(t *testing.T) {
	suite.Run(t, new(RedisSuite))
}
