package request

type Product struct {
	Name        string `json:"name" validate:"required"`
	Price       *int64 `json:"price" validate:"required"`
	Description string `json:"description" validate:"required"`
	Quantity    *int   `json:"quantity" validate:"required"`
	CreatedBy   string `json:"created_by" validate:"required"`
}

type ProductFilter struct {
	Search string `json:"search"`
	Sort   string `json:"sort"`
	Page   int    `json:"page"`
	Limit  int    `json:"limit"`
}
