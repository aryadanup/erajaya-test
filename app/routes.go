package app

import (
	"context"
	"erajaya-test/internal/delivery/http"
	"erajaya-test/internal/repository"
	"erajaya-test/internal/usecase"
	"erajaya-test/shared/response"

	"github.com/labstack/echo/v4"
)

func InitRoutes(ctx context.Context, apiGroup *echo.Group, db *Database) {

	zapLogger := InitZapLogger()

	stdResponse := response.NewStdResponse(zapLogger)

	productRepository := repository.NewProductRepository(db.Postgres)
	productRedis := repository.NewRedisRepository(db.Redis)
	productUsecase := usecase.NewProductUsecase(productRepository, productRedis)
	productHandler := http.NewHandler(productUsecase, stdResponse)

	v1 := apiGroup.Group("/v1")

	v1.POST("/products", productHandler.CreateProduct)
	v1.GET("/products", productHandler.ListProducts)
	v1.GET("/products/:id", productHandler.GetProductByID)

}
