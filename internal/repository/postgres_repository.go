package repository

import (
	"context"

	"erajaya-test/internal/interfaces"
	"erajaya-test/internal/models/entity"
	"erajaya-test/internal/models/request"
	"erajaya-test/shared/constant"

	"gorm.io/gorm"
)

type productRepository struct {
	db *gorm.DB
}

func NewProductRepository(db *gorm.DB) interfaces.ProductRepository {
	return &productRepository{
		db: db,
	}
}

func (r *productRepository) Create(ctx context.Context, product *entity.Product) error {
	return r.db.WithContext(ctx).Create(product).Error
}

func (r *productRepository) GetByID(ctx context.Context, id int64) (*entity.Product, error) {
	var product entity.Product
	err := r.db.WithContext(ctx).First(&product, id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, constant.ErrNotFound
		}
		return nil, err
	}
	return &product, nil
}

func (r *productRepository) Fetch(ctx context.Context, filter request.ProductFilter) ([]entity.Product, int64, error) {
	var products []entity.Product
	var total int64

	query := r.db.Model(&entity.Product{})

	if filter.Search != "" {
		search := "%" + filter.Search + "%"
		query = query.Where("name ILIKE ? OR description ILIKE ?", search, search)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	switch filter.Sort {
	case "newest":
		query = query.Order("created_at DESC")
	case "cheapest":
		query = query.Order("price ASC")
	case "expensive":
		query = query.Order("price DESC")
	case "name asc":
		query = query.Order("name ASC")
	case "name desc":
		query = query.Order("name DESC")
	default:
		query = query.Order("created_at DESC")
	}

	offset := (filter.Page - 1) * filter.Limit
	query = query.Offset(offset).Limit(filter.Limit)

	if err := query.Find(&products).Error; err != nil {
		return nil, 0, err
	}

	return products, total, nil
}
