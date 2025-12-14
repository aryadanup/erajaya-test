package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"erajaya-test/internal/interfaces"
	"erajaya-test/internal/models/entity"
	"erajaya-test/internal/models/request"
	"erajaya-test/internal/repository"
	"erajaya-test/shared/constant"
	"erajaya-test/shared/response"

	"github.com/google/go-querystring/query"
)

type productUsecase struct {
	repo      interfaces.ProductRepository
	redisRepo repository.RedisRepository
}

func NewProductUsecase(repo interfaces.ProductRepository, redisRepo repository.RedisRepository) interfaces.ProductUsecase {
	return &productUsecase{
		repo:      repo,
		redisRepo: redisRepo,
	}
}

func (u *productUsecase) CreateProduct(ctx context.Context, req *request.Product) error {

	product := &entity.Product{
		Name:        req.Name,
		Price:       req.Price,
		Description: req.Description,
		Quantity:    req.Quantity,
		CreatedBy:   req.CreatedBy,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	err := u.repo.Create(ctx, product)
	if err != nil {
		return err
	}

	_ = u.redisRepo.Delete(ctx, "products*")

	return nil
}

func (u *productUsecase) GetProductByID(ctx context.Context, id int64) (*entity.Product, error) {

	key := fmt.Sprintf("%s:%d", constant.RedisKeyProductDetail, id)

	val, err := u.redisRepo.Get(ctx, key)
	if err == nil {
		var product entity.Product
		if err := json.Unmarshal([]byte(val), &product); err == nil {
			return &product, nil
		}
	}

	product, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	data, _ := json.Marshal(product)
	_ = u.redisRepo.Set(ctx, key, data, 5*time.Minute)

	return product, nil
}

func (u *productUsecase) ListProducts(ctx context.Context, filter request.ProductFilter) ([]entity.Product, response.StdPagination, error) {

	query, _ := query.Values(filter)
	queryString := query.Encode()

	key := fmt.Sprintf("%s:%s", constant.RedisKeyProductList, queryString)

	val, err := u.redisRepo.Get(ctx, key)
	if err == nil {
		var result entity.FetchResult
		if err := json.Unmarshal([]byte(val), &result); err == nil {
			pagination := response.StandardPagination(filter.Page, filter.Limit, result.Total)
			return result.Products, pagination, nil
		}
	}

	products, total, err := u.repo.Fetch(ctx, filter)
	if err != nil {
		return nil, response.StdPagination{}, err
	}

	pagination := response.StandardPagination(filter.Page, filter.Limit, total)

	if len(products) > 0 {
		result := entity.FetchResult{
			Products: products,
			Total:    total,
		}
		data, _ := json.Marshal(result)
		_ = u.redisRepo.Set(ctx, key, data, 5*time.Minute)
	}
	return products, pagination, nil
}
