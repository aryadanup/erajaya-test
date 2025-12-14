package entity

import (
	"time"
)

type Product struct {
	ID          int64     `json:"id" gorm:"primaryKey;autoIncrement" readonly:"true"`
	Name        string    `json:"name" gorm:"index:idx_product_name;not null"`
	Price       *int64    `json:"price" gorm:"index:idx_product_price;not null"`
	Description string    `json:"description"`
	Quantity    *int      `json:"quantity"`
	CreatedAt   time.Time `json:"created_at" gorm:"index:idx_product_created_at"`
	CreatedBy   string    `json:"created_by"`
	UpdatedAt   time.Time `json:"updated_at"`
	UpdatedBy   string    `json:"updated_by"`
	DeletedAt   time.Time `json:"deleted_at"`
	DeletedBy   string    `json:"deleted_by"`
}

func (Product) TableName() string {
	return "products"
}

type FetchResult struct {
	Products []Product `json:"products"`
	Total    int64     `json:"total"`
}
