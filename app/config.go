package app

import (
	"fmt"
	"log"
	"strings"

	"erajaya-test/swagger"

	"github.com/labstack/echo/v4"
	"github.com/spf13/viper"
	echoSwagger "github.com/swaggo/echo-swagger"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
}

type ServerConfig struct {
	Port    string
	Timeout int
	Debug   bool
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

func InitConfig(path string) (*Config, error) {

	viper.SetDefault("server.host", "localhost")
	viper.SetDefault("server.port", "8080")
	viper.SetDefault("server.timeout", 30)
	viper.SetDefault("server.debug", false)

	viper.AddConfigPath(path)
	viper.SetConfigName("config")
	viper.SetConfigType("json")

	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
		} else {
			return nil, err
		}
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

type SwaggerInfo struct {
	Name        string
	Env         string
	Host        string
	Port        string
	Version     string
	Description string
}

func InitSwagger(e *echo.Echo, swaggerInfo SwaggerInfo) {

	if swaggerInfo.Env == "production" {
		return
	}

	swagger.SwaggerInfo.Title = swaggerInfo.Name
	swagger.SwaggerInfo.Description = swaggerInfo.Description

	swagger.SwaggerInfo.Host = fmt.Sprintf("%s:%s", swaggerInfo.Host, swaggerInfo.Port)

	if swaggerInfo.Version == "" {
		swaggerInfo.Version = "1.0.0"
	}
	swagger.SwaggerInfo.Version = swaggerInfo.Version

	e.GET("/swagger/*", echoSwagger.WrapHandler)

	log.Printf("[Swagger] Enabled: http://%s:%s/swagger/index.html", swaggerInfo.Host, swaggerInfo.Port)
}
