package http_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"erajaya-test/app"
	productHttp "erajaya-test/internal/delivery/http"
	"erajaya-test/internal/models/entity"
	"erajaya-test/internal/models/request"
	"erajaya-test/mocks"

	"erajaya-test/shared/constant"
	"erajaya-test/shared/response"
	"erajaya-test/shared/utils"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type CustomValidator struct {
	validator *validator.Validate
}

func (cv *CustomValidator) Validate(i interface{}) error {
	return cv.validator.Struct(i)
}

type FailValidator struct{}

func (fv *FailValidator) Validate(i interface{}) error {
	return errors.New("generic validation error")
}

type ProductHandlerTestSuite struct {
	suite.Suite
	echo     *echo.Echo
	mockUC   *mocks.ProductUsecase
	handler  *productHttp.ProductHandler
	recorder *httptest.ResponseRecorder
}

func (s *ProductHandlerTestSuite) SetupTest() {

	s.echo = echo.New()
	s.echo.Validator = &CustomValidator{validator: validator.New()}

	s.mockUC = new(mocks.ProductUsecase)

	logger := app.InitZapLogger()
	resp := response.NewStdResponse(logger)
	s.handler = productHttp.NewHandler(s.mockUC, resp)

	s.recorder = httptest.NewRecorder()
}

func (s *ProductHandlerTestSuite) sendRequest(method, path, body string) echo.Context {
	var req *http.Request
	if body != "" {
		req = httptest.NewRequest(method, path, strings.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	} else {
		req = httptest.NewRequest(method, path, nil)
	}

	s.recorder = httptest.NewRecorder()
	return s.echo.NewContext(req, s.recorder)
}

func (s *ProductHandlerTestSuite) TestCreateProduct() {
	reqJSON := `{"name":"LG TV 42 Inch","price":5000000,"description":"LG TV 42 Inch Full HD","quantity":10,"created_by":"arya"}`

	s.Run("Success", func() {
		c := s.sendRequest(http.MethodPost, "/products", reqJSON)

		s.mockUC.On("CreateProduct", mock.Anything, mock.MatchedBy(func(p *request.Product) bool {
			return p.Name == "LG TV 42 Inch" && *p.Price == 5000000
		})).Return(nil).Once()

		err := s.handler.CreateProduct(c)

		s.NoError(err)
		s.Equal(http.StatusCreated, s.recorder.Code)
	})

	s.Run("Bad Request - Invalid JSON", func() {
		c := s.sendRequest(http.MethodPost, "/products", "invalid-json")

		err := s.handler.CreateProduct(c)

		s.NoError(err)
		s.Equal(http.StatusBadRequest, s.recorder.Code)
	})

	s.Run("Bind Error - Invalid Type", func() {
		c := s.sendRequest(http.MethodPost, "/products", `{"price": "abc"}`)

		err := s.handler.CreateProduct(c)

		s.NoError(err)
		s.Equal(http.StatusBadRequest, s.recorder.Code)
	})

	s.Run("Validation Error - Multiple Missing Fields", func() {
		incompleteJSON := `{"name":"LG TV Only Name"}`
		c := s.sendRequest(http.MethodPost, "/products", incompleteJSON)

		err := s.handler.CreateProduct(c)

		s.NoError(err)
		s.Equal(http.StatusBadRequest, s.recorder.Code)

		var resp struct {
			Message string                  `json:"message"`
			Code    string                  `json:"code"`
			Error   []utils.ValidationError `json:"error"`
		}

		errJSON := json.Unmarshal(s.recorder.Body.Bytes(), &resp)
		s.NoError(errJSON)

		expectedErrors := []utils.ValidationError{
			{Parameter: "Price is required"},
			{Parameter: "Description is required"},
			{Parameter: "Quantity is required"},
			{Parameter: "CreatedBy is required"},
		}

		s.ElementsMatch(expectedErrors, resp.Error, "Error list harus mencakup semua field yang missing")
	})

	s.Run("Usecase Error", func() {
		c := s.sendRequest(http.MethodPost, "/products", reqJSON)

		s.mockUC.On("CreateProduct", mock.Anything, mock.Anything).Return(errors.New(constant.ErrInternal.Error())).Once()

		err := s.handler.CreateProduct(c)

		s.NoError(err)
		s.Equal(http.StatusInternalServerError, s.recorder.Code)
	})
}

func (s *ProductHandlerTestSuite) TestListProducts() {
	s.Run("Success", func() {
		c := s.sendRequest(http.MethodGet, "/products?search=test&sort=newest&page=1&limit=10", "")

		s.mockUC.On("ListProducts", mock.Anything, mock.MatchedBy(func(f request.ProductFilter) bool {
			return f.Search == "test" && f.Sort == "newest" && f.Page == 1 && f.Limit == 10
		})).Return([]entity.Product{}, response.StdPagination{Page: 1, Limit: 10}, nil).Once()

		err := s.handler.ListProducts(c)

		s.NoError(err)
		s.Equal(http.StatusOK, s.recorder.Code)
	})

	s.Run("Success with Default Pagination", func() {

		c := s.sendRequest(http.MethodGet, "/products", "")

		s.mockUC.On("ListProducts", mock.Anything, mock.MatchedBy(func(f request.ProductFilter) bool {
			return f.Page == 1 && f.Limit == 10
		})).Return([]entity.Product{}, response.StdPagination{}, nil).Once()

		err := s.handler.ListProducts(c)
		s.NoError(err)
		s.Equal(http.StatusOK, s.recorder.Code)
	})

	s.Run("Usecase Error", func() {
		c := s.sendRequest(http.MethodGet, "/products", "")

		s.mockUC.On("ListProducts", mock.Anything, mock.Anything).
			Return([]entity.Product{}, response.StdPagination{}, errors.New("db error")).Once()

		err := s.handler.ListProducts(c)

		s.NoError(err)
		s.Equal(http.StatusInternalServerError, s.recorder.Code)
	})
}

func (s *ProductHandlerTestSuite) TestGetProductByID() {
	s.Run("Success", func() {
		c := s.sendRequest(http.MethodGet, "/products/1", "")
		c.SetPath("/products/:id")
		c.SetParamNames("id")
		c.SetParamValues("1")

		expected := &entity.Product{ID: 1, Name: "Test"}
		s.mockUC.On("GetProductByID", mock.Anything, int64(1)).Return(expected, nil).Once()

		err := s.handler.GetProductByID(c)

		s.NoError(err)
		s.Equal(http.StatusOK, s.recorder.Code)
	})

	s.Run("Invalid ID (Parse Error)", func() {

		c := s.sendRequest(http.MethodGet, "/products/abc", "")
		c.SetPath("/products/:id")
		c.SetParamNames("id")
		c.SetParamValues("abc")

		s.mockUC.On("GetProductByID", mock.Anything, int64(0)).Return(nil, constant.ErrNotFound).Once()

		err := s.handler.GetProductByID(c)

		s.NoError(err)
		s.Equal(http.StatusNotFound, s.recorder.Code)
	})

	s.Run("Not Found", func() {
		c := s.sendRequest(http.MethodGet, "/products/999", "")
		c.SetPath("/products/:id")
		c.SetParamNames("id")
		c.SetParamValues("999")

		s.mockUC.On("GetProductByID", mock.Anything, int64(999)).Return(nil, constant.ErrNotFound).Once()

		err := s.handler.GetProductByID(c)

		s.NoError(err)
		s.Equal(http.StatusNotFound, s.recorder.Code)
	})

	s.Run("Usecase Error", func() {
		c := s.sendRequest(http.MethodGet, "/products/1", "")
		c.SetPath("/products/:id")
		c.SetParamNames("id")
		c.SetParamValues("1")

		s.mockUC.On("GetProductByID", mock.Anything, int64(1)).Return(nil, errors.New("db error")).Once()

		err := s.handler.GetProductByID(c)

		s.NoError(err)
		s.Equal(http.StatusInternalServerError, s.recorder.Code)
	})
}

func TestProductHandlerSuite(t *testing.T) {
	suite.Run(t, new(ProductHandlerTestSuite))
}
