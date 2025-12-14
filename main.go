package main

import (
	"context"
	"strings"

	"erajaya-test/app"
	"erajaya-test/shared/middlewares"
	"erajaya-test/shared/response"
	"erajaya-test/shared/utils"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/spf13/viper"
	"golang.org/x/time/rate"
)

// @title erajaya-test Product API
// @version 1.0
// @description Product management API
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.email support@erajaya-test.com

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /
// @schemes http

func main() {

	_, err := app.InitConfig("./conf/")
	if err != nil {
		fmt.Println(err)
	}

	env := strings.ToLower(viper.GetString("server.env"))
	version := viper.GetString("server.version")
	appName := viper.GetString("server.app_name")
	port := viper.GetString("server.port")

	initTimeout := viper.GetInt("server.timeout")
	if initTimeout == 0 {
		initTimeout = 30
	}
	initCtx, initCancel := context.WithTimeout(context.Background(), time.Duration(initTimeout)*time.Second)
	defer initCancel()

	e := echo.New()

	e.Validator = utils.NewValidator()

	e.Use(middleware.Recover())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete},
		AllowHeaders: []string{"Origin", "Content-Type", "Accept", "Authorization"},
	}))

	e.IPExtractor = echo.ExtractIPFromXFFHeader()

	rateLimit := viper.GetInt("server.rate_limit")
	if rateLimit <= 0 {
		rateLimit = 20
	}

	rateLimitConfig := middleware.RateLimiterConfig{
		Skipper: middleware.DefaultSkipper,
		Store:   middleware.NewRateLimiterMemoryStore(rate.Limit(rateLimit)),
		IdentifierExtractor: func(ctx echo.Context) (string, error) {
			realIP := ctx.RealIP()
			return realIP, nil
		},

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

	e.Use(middleware.RateLimiterWithConfig(rateLimitConfig))
	e.Use(middleware.TimeoutWithConfig(middleware.TimeoutConfig{
		Timeout: 60 * time.Second,
		OnTimeoutRouteErrorHandler: func(err error, c echo.Context) {
			c.JSON(http.StatusRequestTimeout, map[string]interface{}{
				"error":   err.Error(),
				"code":    response.CodeRequestTimeout,
				"data":    nil,
				"message": response.RequestTimeout,
			})
		},
	}))
	e.Use(middleware.RemoveTrailingSlash())
	ctxMiddleware := middlewares.DefaultCtx{
		BaseUrl: fmt.Sprintf("http://%s:%s", viper.GetString("server.host"), viper.GetString("server.port")),
	}
	e.Use(ctxMiddleware.ContextMiddleware())

	api := e.Group("/api")

	dbInstance := app.InitDatabase(initCtx)
	app.InitRoutes(initCtx, api, dbInstance)

	host := viper.GetString("server.host")

	app.InitSwagger(e, app.SwaggerInfo{
		Name:        appName,
		Env:         env,
		Host:        host,
		Port:        port,
		Version:     version,
		Description: "Product management API with multiple database support",
	})

	go func() {
		log.Printf("Server starting on port %s", port)

		if err := e.Start(":" + port); err != nil && err != http.ErrServerClosed {
			e.Logger.Fatal("shutting down the server caused by: ", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	<-quit
	log.Println("Shutting down server...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := e.Shutdown(shutdownCtx); err != nil {
		e.Logger.Fatal(err)
	}

	// Close all database connections
	dbInstance.Close(shutdownCtx)

	log.Println("Server exited gracefully")
}
