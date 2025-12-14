package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"
	"time"

	"erajaya-test/internal/interfaces"
	"erajaya-test/internal/models/entity"
	"erajaya-test/internal/models/request"
	"erajaya-test/mocks"
	"erajaya-test/shared/constant"

	"github.com/google/go-querystring/query"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type ProductUsecaseTestSuite struct {
	suite.Suite
	mockRepo      *mocks.ProductRepository
	mockRedisRepo *mocks.RedisRepository
	uc            interfaces.ProductUsecase
}

func (s *ProductUsecaseTestSuite) SetupTest() {
	s.mockRepo = new(mocks.ProductRepository)
	s.mockRedisRepo = new(mocks.RedisRepository)
	s.uc = NewProductUsecase(s.mockRepo, s.mockRedisRepo)
}

func (s *ProductUsecaseTestSuite) TestCreateProduct() {
	price := int64(5000000)
	qty := 10
	req := &request.Product{
		Name:        "LG TV",
		Price:       &price,
		Description: "Desc",
		Quantity:    &qty,
		CreatedBy:   "arya",
	}

	s.Run("Success", func() {

		s.mockRepo.On("Create", mock.Anything, mock.MatchedBy(func(p *entity.Product) bool {
			return p.Name == "LG TV" && p.CreatedBy == "arya"
		})).Return(nil).Once()

		s.mockRedisRepo.On("Delete", mock.Anything, "products*").Return(nil).Once()

		err := s.uc.CreateProduct(context.Background(), req)

		s.NoError(err)
	})

	s.Run("Repository Error", func() {
		s.mockRepo.On("Create", mock.Anything, mock.Anything).Return(errors.New("db error")).Once()

		err := s.uc.CreateProduct(context.Background(), req)

		s.Error(err)
		s.Equal("db error", err.Error())
	})
}

func (s *ProductUsecaseTestSuite) TestGetProductByID() {
	id := int64(1)
	key := fmt.Sprintf("%s:%d", constant.RedisKeyProductDetail, id)
	mockProduct := &entity.Product{ID: id, Name: "Test Product"}

	s.Run("Cache Hit (Data found in Redis)", func() {

		dataBytes, _ := json.Marshal(mockProduct)

		s.mockRedisRepo.On("Get", mock.Anything, key).Return(string(dataBytes), nil).Once()

		result, err := s.uc.GetProductByID(context.Background(), id)

		s.NoError(err)
		s.Equal(mockProduct.ID, result.ID)
		s.Equal(mockProduct.Name, result.Name)
	})

	s.Run("Cache Miss (Data fetch from DB)", func() {
		s.mockRedisRepo.On("Get", mock.Anything, key).Return("", errors.New("redis: nil")).Once()

		s.mockRepo.On("GetByID", mock.Anything, id).Return(mockProduct, nil).Once()

		s.mockRedisRepo.On("Set", mock.Anything, key, mock.Anything, 5*time.Minute).Return(nil).Once()

		result, err := s.uc.GetProductByID(context.Background(), id)

		s.NoError(err)
		s.Equal(mockProduct.ID, result.ID)
	})

	s.Run("Repository Error", func() {
		s.mockRedisRepo.On("Get", mock.Anything, key).Return("", errors.New("redis: nil")).Once()
		s.mockRepo.On("GetByID", mock.Anything, id).Return(nil, errors.New("db error")).Once()

		result, err := s.uc.GetProductByID(context.Background(), id)

		s.Error(err)
		s.Nil(result)
	})
}

func (s *ProductUsecaseTestSuite) TestListProducts() {
	filter := request.ProductFilter{
		Search: "LG",
		Sort:   "newest",
		Page:   1,
		Limit:  10,
	}

	v, _ := query.Values(filter)
	expectedKey := fmt.Sprintf("%s:%s", constant.RedisKeyProductList, v.Encode())

	mockProducts := []entity.Product{
		{ID: 1, Name: "LG TV"},
	}
	mockTotal := int64(1)

	s.Run("Cache Hit (Return from Redis)", func() {

		cachedData := entity.FetchResult{
			Products: mockProducts,
			Total:    mockTotal,
		}
		dataBytes, _ := json.Marshal(cachedData)

		s.mockRedisRepo.On("Get", mock.Anything, expectedKey).Return(string(dataBytes), nil).Once()

		results, pagination, err := s.uc.ListProducts(context.Background(), filter)

		s.NoError(err)
		s.Equal(1, len(results))
		s.Equal(int(mockTotal), pagination.Total)
	})

	s.Run("Cache Miss - Data Found (Save to Redis)", func() {

		s.mockRedisRepo.On("Get", mock.Anything, expectedKey).Return("", errors.New("redis: nil")).Once()

		s.mockRepo.On("Fetch", mock.Anything, filter).Return(mockProducts, mockTotal, nil).Once()

		s.mockRedisRepo.On("Set", mock.Anything, expectedKey, mock.Anything, 5*time.Minute).Return(nil).Once()

		results, _, err := s.uc.ListProducts(context.Background(), filter)

		s.NoError(err)
		s.Equal(mockProducts[0].Name, results[0].Name)
	})

	s.Run("Cache Miss - No Data (Skip Save Redis)", func() {

		s.mockRedisRepo.On("Get", mock.Anything, expectedKey).Return("", errors.New("redis: nil")).Once()

		s.mockRepo.On("Fetch", mock.Anything, filter).Return([]entity.Product{}, int64(0), nil).Once()

		results, pagination, err := s.uc.ListProducts(context.Background(), filter)

		s.NoError(err)
		s.Empty(results)
		s.Equal(0, pagination.Total)
	})

	s.Run("Repository Error", func() {
		s.mockRedisRepo.On("Get", mock.Anything, expectedKey).Return("", errors.New("redis: nil")).Once()
		s.mockRepo.On("Fetch", mock.Anything, filter).Return(nil, int64(0), errors.New("db error")).Once()

		results, _, err := s.uc.ListProducts(context.Background(), filter)

		s.Error(err)
		s.Nil(results)
	})
}

func TestProductUsecaseSuite(t *testing.T) {
	suite.Run(t, new(ProductUsecaseTestSuite))
}
