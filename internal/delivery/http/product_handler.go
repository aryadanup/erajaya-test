package http

import (
	"erajaya-test/internal/interfaces"
	"erajaya-test/internal/models/request"
	"erajaya-test/shared/constant"
	"erajaya-test/shared/response"
	"strconv"

	"github.com/labstack/echo/v4"
)

type ProductHandler struct {
	usecase  interfaces.ProductUsecase
	response *response.StdResponse
}

func NewHandler(productUsecase interfaces.ProductUsecase, standardResponse *response.StdResponse) *ProductHandler {
	return &ProductHandler{
		usecase:  productUsecase,
		response: standardResponse,
	}
}

// CreateProduct godoc
// @Summary Create a new product
// @Description Create a new product with the provided information
// @Tags products
// @Accept json
// @Produce json
// @Param product body request.Product true "Product object"
// @Success 201 {object} response.StdResponse
// @Failure 400 {object} response.StdResponse
// @Failure 500 {object} response.StdResponse
// @Router /api/v1/products [post]
func (h *ProductHandler) CreateProduct(c echo.Context) error {
	var req request.Product
	if err := c.Bind(&req); err != nil {
		return h.response.StandardResponse(c, h.response.ErrorResponse(c.Request().Context(), response.BadRequest, err, "PRD-ERA-410"))
	}

	if err := c.Validate(&req); err != nil {
		return h.response.StandardResponse(c, h.response.ErrorResponse(c.Request().Context(), response.BadRequest, err, "PRD-ERA-400"))
	}

	ctx := c.Request().Context()
	err := h.usecase.CreateProduct(ctx, &req)
	if err != nil {
		return h.response.StandardResponse(c, h.response.ErrorResponse(ctx, response.InternalError, err, "PRD-ERA-500"))
	}

	return h.response.StandardResponse(c, h.response.SuccessResponse(ctx, response.InsertSuccess, req, "PRD-ERA-201"))
}

// ListProducts godoc
// @Summary List all products
// @Description Get a list of products with optional filtering and pagination
// @Tags products
// @Accept json
// @Produce json
// @Param search query string false "Search term"
// @Param sort query string false "Sort field"
// @Param page query int false "Page number"
// @Param limit query int false "Items per page"
// @Success 200 {object} response.StdResponse
// @Failure 500 {object} response.StdResponse
// @Router /api/v1/products [get]
func (h *ProductHandler) ListProducts(c echo.Context) error {
	search := c.QueryParam("search")
	sort := c.QueryParam("sort")
	page, _ := strconv.Atoi(c.QueryParam("page"))
	limit, _ := strconv.Atoi(c.QueryParam("limit"))

	filter := request.ProductFilter{
		Search: search,
		Sort:   sort,
		Page:   page,
		Limit:  limit,
	}

	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.Limit <= 0 {
		filter.Limit = 10
	}

	ctx := c.Request().Context()
	products, metadata, err := h.usecase.ListProducts(ctx, filter)
	if err != nil {
		return h.response.StandardResponse(c, h.response.ErrorResponse(ctx, response.InternalError, err, "PRD-ERA-500"))
	}

	return h.response.StandardResponse(c, h.response.SuccessResponse(ctx, response.GetSuccess, map[string]interface{}{
		"data":     products,
		"metadata": metadata,
	}, "PRD-ERA-200"))
}

// GetProductByID godoc
// @Summary Get product by ID
// @Description Get a single product by its ID
// @Tags products
// @Accept json
// @Produce json
// @Param id path int true "Product ID"
// @Success 200 {object} response.StdResponse
// @Failure 404 {object} response.StdResponse
// @Failure 500 {object} response.StdResponse
// @Router /api/v1/products/{id} [get]
func (h *ProductHandler) GetProductByID(c echo.Context) error {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	ctx := c.Request().Context()
	product, err := h.usecase.GetProductByID(ctx, id)
	if err != nil {
		if err == constant.ErrNotFound {
			return h.response.StandardResponse(c, h.response.ErrorResponse(ctx, response.NotFound, err, "PRD-ERA-404"))
		}
		return h.response.StandardResponse(c, h.response.ErrorResponse(ctx, response.InternalError, err, "PRD-ERA-500"))
	}

	return h.response.StandardResponse(c, h.response.SuccessResponse(ctx, response.GetSuccess, product, "PRD-ERA-200"))
}
