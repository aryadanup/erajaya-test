package constant

import "errors"

var (
	// Error
	ErrValidation = errors.New("validation error")
	ErrInternal   = errors.New("internal server error")
	ErrNotFound   = errors.New("record not found")

	// Redis Key
	RedisKeyProductDetail = "products:detail"
	RedisKeyProductList   = "products:list"
)
