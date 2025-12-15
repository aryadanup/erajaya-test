package response

import (
	"context"
	"erajaya-test/shared/middlewares"
	"erajaya-test/shared/utils"
	"net/http"
	"reflect"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type ApiResponse struct {
	Message  any    `json:"message"`
	Data     any    `json:"data,omitempty"`
	Error    any    `json:"error,omitempty"`
	Code     string `json:"code"`
	Metadata any    `json:"metadata,omitempty"`
	HTTPCode any    `json:"-"`
}

type StdPagination struct {
	NextPage bool `json:"next_page"`
	PrevPage bool `json:"prev_page"`
	Limit    int  `json:"limit"`
	Page     int  `json:"page"`
	Total    int  `json:"total"`
}

type StdMessage string

const (
	InsertSuccess    StdMessage = "data successfully inserted"
	GetSuccess       StdMessage = "data successfully retrieved"
	BadRequest       StdMessage = "your data validation is incorrect please check again"
	NotFound         StdMessage = "data not found"
	MethodNotAllowed StdMessage = "method not allowed"
	RequestTimeout   StdMessage = "the request has exceeded the time limit please try again"
	TooManyRequests  StdMessage = "too many requests please try again in a moment"
	InternalError    StdMessage = "internal server error"
)

const (
	CodeBadRequest          = "PRD-ERA-400"
	CodeErrorBind           = "PRD-ERA-410"
	CodeSuccess             = "PRD-ERA-200"
	CodeCreated             = "PRD-ERA-201"
	CodeNotFound            = "PRD-ERA-404"
	CodeMethodNotAllowed    = "PRD-ERA-405"
	CodeRequestTimeout      = "PRD-ERA-408"
	CodeTooManyRequests     = "PRD-ERA-429"
	CodeInternalServerError = "PRD-ERA-500"
)

type StdResponse struct {
	zapLogger *zap.Logger
}

func NewStdResponse(zapLogger *zap.Logger) *StdResponse {
	return &StdResponse{
		zapLogger: zapLogger,
	}
}

type JSONResponse interface {
	SuccessResponse(ctx context.Context, message StdMessage, data any, code string) *ApiResponse
	ErrorResponse(ctx context.Context, message StdMessage, error error, code string) *ApiResponse
	StandardResponse(ctx echo.Context, response *ApiResponse) error
}

func (s *StdResponse) SuccessResponse(ctx context.Context, message StdMessage, data any, code string) *ApiResponse {

	switch message {
	case InsertSuccess:
		return &ApiResponse{
			Message:  message,
			Data:     data,
			Code:     code,
			HTTPCode: http.StatusCreated,
		}
	default:
		return &ApiResponse{
			Message:  message,
			Data:     data,
			Code:     code,
			HTTPCode: http.StatusOK,
		}
	}
}

func (s *StdResponse) ErrorResponse(ctx context.Context, message StdMessage, err error, code string) *ApiResponse {

	switch message {

	case BadRequest:
		validationErrors := utils.ParseValidationErrors(err)

		var formattedError interface{}

		if len(validationErrors) > 0 {
			formattedError = validationErrors
		} else {
			if err != nil {
				formattedError = []utils.ValidationError{
					{Parameter: err.Error()},
				}
			}
		}

		return &ApiResponse{
			Message:  message,
			Error:    formattedError,
			Code:     code,
			HTTPCode: http.StatusBadRequest,
		}

	case NotFound:
		return &ApiResponse{
			Message:  message,
			Error:    err.Error(),
			Code:     code,
			HTTPCode: http.StatusNotFound,
		}
	case MethodNotAllowed:
		return &ApiResponse{
			Message:  message,
			Error:    err.Error(),
			Code:     code,
			HTTPCode: http.StatusMethodNotAllowed,
		}
	case RequestTimeout:
		return &ApiResponse{
			Message:  message,
			Error:    err.Error(),
			Code:     code,
			HTTPCode: http.StatusRequestTimeout,
		}
	case TooManyRequests:
		return &ApiResponse{
			Message:  message,
			Error:    err.Error(),
			Code:     code,
			HTTPCode: http.StatusTooManyRequests,
		}
	default:
		return &ApiResponse{
			Message:  message,
			Error:    err.Error(),
			Code:     code,
			HTTPCode: http.StatusInternalServerError,
		}
	}

}

func (s *StdResponse) StandardResponse(ctx echo.Context, response *ApiResponse) error {

	c := ctx.Request().Context()

	var xRequestId string
	if xReqId, ok := c.Value(middlewares.CtxRequestID).(string); ok {
		xRequestId = xReqId
	} else {
		xRequestId = ctx.Request().Header.Get(echo.HeaderXRequestID)
	}

	var duration interface{} = int64(0)
	if startTime, ok := c.Value(middlewares.CtxRequestTime).(int64); ok {
		duration = calcDuration(startTime)
	}

	var bodyPayload interface{}
	if v := c.Value(middlewares.CtxRequestPayload); v != nil {
		bodyPayload = structToMap(v)
	}

	zapField := []zap.Field{
		zap.String(echo.HeaderXRequestID, xRequestId),
		zap.String("Method", ctx.Request().Method),
		zap.Any("Header", setHeader(cleanHeader(ctx.Request().Header))),
		zap.Any("Duration", duration),
		zap.Any("Body", bodyPayload),
		zap.Any("Url", c.Value(middlewares.CtxIncomingRequestURL)),
		zap.String("ServerTime", time.Now().Local().Format("2006/01/02 15:04:05.000")),
	}

	if response.Error != nil {
		s.zapLogger.Error("response error",
			zapField...,
		)
	} else {
		s.zapLogger.Info("response success",
			zapField...,
		)
	}

	httpCode := setEmptyHTTPCode(response)

	return ctx.JSON(httpCode, response)
}

func setEmptyHTTPCode(response *ApiResponse) int {
	httpCode := response.HTTPCode.(int)
	response.HTTPCode = nil
	return httpCode
}

func calcDuration(startTime int64) float64 {
	if startTime == 0 {
		return 0
	}
	return float64(time.Now().Local().UnixMilli() - startTime)
}

func cleanHeader(h http.Header) http.Header {
	clean := make(http.Header)
	for k, v := range h {
		switch strings.ToLower(k) {
		case "accept", "accept-encoding", "accept-language", "connection", "sec-ch-ua", "sec-ch-ua-mobile",
			"sec-ch-ua-platform", "sec-fetch-dest", "sec-fetch-mode", "sec-fetch-site", "sec-fetch-user",
			"upgrade-insecure-requests", "user-agent":
			continue
		default:
			clean[k] = v
		}
	}
	return clean
}

func setHeader(httpHeader http.Header) map[string]interface{} {
	headerMap := make(map[string]interface{})
	for key, values := range httpHeader {
		if len(values) > 0 {
			headerMap[key] = values[0]
		}
	}
	return headerMap
}

func structToMap(item interface{}) map[string]interface{} {
	switch item := item.(type) {
	case map[string]interface{}:
		return item
	case nil:
		return map[string]interface{}{}
	}

	res := map[string]interface{}{}
	v := reflect.TypeOf(item)
	if v.Kind() != reflect.Struct && v.Kind() != reflect.Ptr {
		res["data"] = item
		return res
	}

	reflectValue := reflect.ValueOf(item)
	reflectValue = reflect.Indirect(reflectValue)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	for i := 0; i < v.NumField(); i++ {
		tag := v.Field(i).Tag.Get("json")
		if strings.Contains(tag, ",omitempty") {
			tag = strings.TrimSuffix(tag, ",omitempty")
		}
		field := reflectValue.Field(i).Interface()
		if tag != "" && tag != "-" {
			if v.Field(i).Type.Kind() == reflect.Struct {
				res[tag] = structToMap(field)
			} else {
				res[tag] = field
			}
		} else if tag == "" && field != nil &&
			v.Field(i).Type.Kind() == reflect.Struct {
			tempMap := structToMap(field)
			for k, v := range tempMap {
				res[k] = v
			}
		}
	}
	return res
}

func StandardPagination(page, limit int, total int64) StdPagination {
	return StdPagination{
		Page:     page,
		Limit:    limit,
		Total:    int(total),
		NextPage: (page * limit) < int(total),
		PrevPage: page > 1,
	}
}
