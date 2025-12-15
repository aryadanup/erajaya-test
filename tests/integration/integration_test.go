package integration

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"

	"erajaya-test/app"
	productHandler "erajaya-test/internal/delivery/http"
	"erajaya-test/internal/repository"
	"erajaya-test/internal/usecase"

	"encoding/json"
	"erajaya-test/shared/datastore"
	"erajaya-test/shared/response"
	"erajaya-test/shared/utils"
)

type safeWriter struct {
	http.ResponseWriter
	mu       sync.Mutex
	timedOut bool
}

func (w *safeWriter) Write(b []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.timedOut {
		return 0, nil
	}
	return w.ResponseWriter.Write(b)
}

func (w *safeWriter) WriteHeader(statusCode int) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.timedOut {
		return
	}
	w.ResponseWriter.WriteHeader(statusCode)
}

type ProductTestSuite struct {
	suite.Suite
	echo     *echo.Echo
	db       *gorm.DB
	redis    *redis.Client
	cleanups []func()
}

func (s *ProductTestSuite) SetupSuite() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	s.cleanups = append(s.cleanups, cancel)

	// Setup Postgres
	pgCfg := datastore.Config{Host: "localhost", Port: 5432, User: "user", Password: "password", DBName: "erajaya_db"}
	pgFactory, err := datastore.NewDatastoreFactory(datastore.Postgres, pgCfg)
	s.Require().NoError(err)
	s.Require().NoError(pgFactory.Connect(ctx))
	s.db = pgFactory.GetClient().(*gorm.DB)

	// Setup Redis
	redisCfg := datastore.Config{Host: "localhost", Port: 6379, Password: "", DBName: "0"}
	redisFactory, err := datastore.NewDatastoreFactory(datastore.Redis, redisCfg)
	s.Require().NoError(err)
	if err := redisFactory.Connect(ctx); err == nil {
		s.redis = redisFactory.GetClient().(*redis.Client)
	}

	// Setup Application Logic
	productRepository := repository.NewProductRepository(s.db)
	redisRepo := repository.NewRedisRepository(s.redis)
	productUsecase := usecase.NewProductUsecase(productRepository, redisRepo)

	// Setup Echo
	s.echo = echo.New()
	s.echo.Validator = &utils.CustomValidator{Validator: validator.New()}

	logger := app.InitZapLogger()
	h := productHandler.NewHandler(productUsecase, response.NewStdResponse(logger))

	v1 := s.echo.Group("/api/v1")

	v1.POST("/products", h.CreateProduct)
	v1.GET("/products", h.ListProducts)
	v1.GET("/products/:id", h.GetProductByID)

	rateLimitConfig := middleware.RateLimiterConfig{
		Skipper: middleware.DefaultSkipper,
		Store:   middleware.NewRateLimiterMemoryStore(5),
		ErrorHandler: func(context echo.Context, err error) error {
			return context.JSON(http.StatusTooManyRequests, map[string]interface{}{
				"error":   nil,
				"code":    response.CodeTooManyRequests,
				"data":    nil,
				"message": response.TooManyRequests,
			})
		},
		DenyHandler: func(context echo.Context, identifier string, err error) error {
			return context.JSON(http.StatusTooManyRequests, map[string]interface{}{
				"error":   nil,
				"code":    response.CodeTooManyRequests,
				"data":    nil,
				"message": response.TooManyRequests,
			})
		},
	}

	rateLimitGroup := s.echo.Group("/rate-limit")
	rateLimitGroup.Use(middleware.RateLimiterWithConfig(rateLimitConfig))
	rateLimitGroup.GET("", func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	s.cleanups = append(s.cleanups, func() {
		pgFactory.Close(ctx)
		redisFactory.Close(ctx)
	})
}

func (s *ProductTestSuite) TearDownSuite() {
	for _, cleanup := range s.cleanups {
		cleanup()
	}
}

func (s *ProductTestSuite) sendRequest(method, target, body, contentType string) *httptest.ResponseRecorder {
	var req *http.Request
	if body != "" {
		req = httptest.NewRequest(method, target, strings.NewReader(body))
	} else {
		req = httptest.NewRequest(method, target, nil)
	}

	if contentType != "" {
		req.Header.Set(echo.HeaderContentType, contentType)
	} else {
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	}

	rec := httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)
	return rec
}

func (s *ProductTestSuite) TestSQLInjection() {
	s.Run("Search_Parameter", func() {
		testcase := []string{
			"' OR '1'='1",
			"' UNION SELECT NULL--",
			"'; DROP TABLE products--",
			"' OR SLEEP(5)--",
		}

		for _, tc := range testcase {
			s.Run("Payload_"+tc, func() {
				encodedPayload := url.QueryEscape(tc)
				rec := s.sendRequest(http.MethodGet, "/api/v1/products?search="+encodedPayload, "", "application/json")

				s.NotEqual(http.StatusInternalServerError, rec.Code, "Search payload caused 500: %s", tc)
			})
		}
	})

	s.Run("Sort_Parameter", func() {
		testcase := []string{
			"created_at; DROP TABLE products--",
			"(CASE WHEN (1=1) THEN created_at ELSE price END)",
			"id desc, (SELECT sleep(5))",
			"non_existent_column",
		}

		for _, tc := range testcase {
			s.Run("Payload_"+tc, func() {
				encodedPayload := url.QueryEscape(tc)

				rec := s.sendRequest(http.MethodGet, "/api/v1/products?sort="+encodedPayload, "", "application/json")

				s.NotEqual(http.StatusInternalServerError, rec.Code, "Sort payload caused 500: %s", tc)
			})
		}
	})
}
func (s *ProductTestSuite) TestXSS() {
	testcase := []string{
		"<script>alert('XSS')</script>",
		"<img src=x onerror=alert('XSS')>",
	}

	for _, tc := range testcase {
		s.Run("Payload_"+tc, func() {
			body := fmt.Sprintf(`{"name":"%s", "price":1000, "description":"%s", "quantity":1}`, tc, tc)
			rec := s.sendRequest(http.MethodPost, "/api/v1/products", body, "application/json")

			if rec.Code == http.StatusCreated {
				s.NotContains(rec.Body.String(), "<script>", "Response content should be escaped")
			}
		})
	}
}

func (s *ProductTestSuite) TestInputValidation() {
	testcase := []struct {
		name      string
		method    string
		url       string
		body      string
		wantCodes []int
	}{
		{"Invalid Page Param", http.MethodGet, "/api/v1/products?page=abc", "", []int{400, 200}},
		{"Negative Limit", http.MethodGet, "/api/v1/products?limit=-1", "", []int{400, 200}},
		{"Huge Body", http.MethodPost, "/api/v1/products", fmt.Sprintf(`{"name":"%s", "price":10}`, strings.Repeat("A", 10000)), []int{400, 422, 201}},
		{"Negative Price", http.MethodPost, "/api/v1/products", `{"name":"T", "price":-500}`, []int{400, 422}},
	}

	for _, tc := range testcase {
		s.Run(tc.name, func() {
			rec := s.sendRequest(tc.method, tc.url, tc.body, "")
			s.NotEqual(http.StatusInternalServerError, rec.Code)
			if len(tc.wantCodes) > 0 {
				s.Contains(tc.wantCodes, rec.Code)
			}
		})
	}
}

func (s *ProductTestSuite) TestInformationDisclosure() {
	body := `{"name": "Broken JSON", "price":`

	rec := s.sendRequest(http.MethodPost, "/api/v1/products", body, "application/json")

	s.Equal(http.StatusBadRequest, rec.Code)

	responseBody := rec.Body.String()
	s.NotContains(responseBody, ".go:", "Critical: Response leaks source code line number")
	s.NotContains(responseBody, "goroutine", "Critical: Response leaks stack trace")
}

func (s *ProductTestSuite) TestMethod() {

	s.Run("Method Not Allowed", func() {
		rec := s.sendRequest(http.MethodDelete, "/api/v1/products", "", "")

		s.Equal(http.StatusMethodNotAllowed, rec.Code, "Should return 405 for unsupported method")
	})

}

func (s *ProductTestSuite) TestContentType() {

	s.Run("Content Type Enforcement", func() {

		req := httptest.NewRequest(http.MethodPost, "/api/v1/products", strings.NewReader(`{"name":"A"}`))
		req.Header.Set(echo.HeaderContentType, "text/plain")

		rec := httptest.NewRecorder()
		s.echo.ServeHTTP(rec, req)

		s.True(rec.Code >= 400, "Should reject non-JSON Content-Type")
	})
}

func (s *ProductTestSuite) TestProduct() {

	body := `{"name":"LG TV 42 inch","price":5000000,"description":"Full HD","quantity":5, "created_by":"arya"}`
	createRec := s.sendRequest(http.MethodPost, "/api/v1/products", body, "application/json")

	if createRec.Code != http.StatusCreated {
		s.T().Logf("Create failed. Response: %s", createRec.Body.String())
	}

	s.Equal(http.StatusCreated, createRec.Code)

	searchRec := s.sendRequest(http.MethodGet, "/api/v1/products?search=LG", "", "application/json")
	s.Equal(http.StatusOK, searchRec.Code)
	s.Contains(searchRec.Body.String(), "LG TV 42 inch")
}

func (s *ProductTestSuite) TestRateLimit() {
	for i := 0; i < 10; i++ {
		rec := s.sendRequest(http.MethodGet, "/rate-limit", "", "application/json")
		if i < 5 {
			s.NotEqual(http.StatusTooManyRequests, rec.Code, "Request %d should pass", i)
		} else {
			if rec.Code == http.StatusTooManyRequests {
				s.Contains(rec.Body.String(), response.CodeTooManyRequests)
			}
		}
	}
}

func customTimeout(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		writer := &safeWriter{ResponseWriter: c.Response().Writer}
		c.Response().Writer = writer

		ctx, cancel := context.WithTimeout(c.Request().Context(), 100*time.Millisecond)
		defer cancel()

		req := c.Request().WithContext(ctx)
		c.SetRequest(req)

		done := make(chan error, 1)
		go func() {
			done <- next(c)
		}()

		select {
		case <-ctx.Done():
			writer.mu.Lock()
			writer.timedOut = true
			writer.mu.Unlock()

			origWriter := writer.ResponseWriter
			origWriter.Header().Set("Content-Type", "application/json")
			origWriter.WriteHeader(http.StatusRequestTimeout)

			resp := map[string]interface{}{
				"error":   "timeout",
				"code":    response.CodeRequestTimeout,
				"data":    nil,
				"message": response.RequestTimeout,
			}
			jsonBytes, _ := json.Marshal(resp)
			origWriter.Write(jsonBytes)

			c.Response().Committed = true

			return nil
		case err := <-done:
			return err
		}
	}
}

func (s *ProductTestSuite) TestTimeout() {

	isolatedEcho := echo.New()
	isolatedEcho.Validator = &utils.CustomValidator{Validator: validator.New()}

	isolatedEcho.GET("/test-timeout", func(c echo.Context) error {
		time.Sleep(300 * time.Millisecond)
		return nil
	}, customTimeout)

	req := httptest.NewRequest(http.MethodGet, "/test-timeout", nil)
	req.Header.Set(echo.HeaderContentType, "application/json")
	rec := httptest.NewRecorder()

	isolatedEcho.ServeHTTP(rec, req)

	s.Equal(http.StatusRequestTimeout, rec.Code, "Should return 408 Request Timeout on timeout")
	s.Contains(rec.Body.String(), response.CodeRequestTimeout)
}

func TestProductSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	suite.Run(t, new(ProductTestSuite))
}
