package interfaces

import (
	"context"
	"erajaya-test/internal/models/entity"
	"erajaya-test/internal/models/request"
	"erajaya-test/shared/response"
)

type ProductRepository interface {
	Create(ctx context.Context, product *entity.Product) error
	GetByID(ctx context.Context, id int64) (*entity.Product, error)
	Fetch(ctx context.Context, filter request.ProductFilter) ([]entity.Product, int64, error)
}

type ProductUsecase interface {
	CreateProduct(ctx context.Context, req *request.Product) error
	GetProductByID(ctx context.Context, id int64) (*entity.Product, error)
	ListProducts(ctx context.Context, filter request.ProductFilter) ([]entity.Product, response.StdPagination, error)
}
