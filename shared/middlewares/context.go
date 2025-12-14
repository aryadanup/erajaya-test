package middlewares

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type contextKey string

const (
	CtxRequestID             contextKey = echo.HeaderXRequestID
	CtxRequestTime           contextKey = "RequestTime"
	CtxIncomingRequestURL    contextKey = "IncomingRequestURL"
	CtxIncomingRequestMethod contextKey = "IncomingRequestMethod"
	CtxRequestPayload        contextKey = "RequestPayload"
)

type DefaultCtx struct {
	BaseUrl string
}

func (m *DefaultCtx) ContextMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			startTime := time.Now().Local().UnixMilli()
			requestId := c.Request().Header.Get(echo.HeaderXRequestID)
			if requestId == "" {
				requestId = uuid.New().String()
			}

			c.Response().Header().Set(echo.HeaderXRequestID, requestId)

			incomingRequestURL := m.BaseUrl + c.Request().URL.Path

			ctx := c.Request().Context()

			if c.Request().Body != nil {
				bodyBytes, err := io.ReadAll(c.Request().Body)
				if err != nil {
					return err
				}

				c.Request().Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

				var body map[string]interface{}
				if err := json.Unmarshal(bodyBytes, &body); err == nil {
					ctx = context.WithValue(ctx, CtxRequestPayload, body)
				}
			}

			ctx = context.WithValue(ctx, CtxRequestID, requestId)
			ctx = context.WithValue(ctx, CtxRequestTime, startTime)
			ctx = context.WithValue(ctx, CtxIncomingRequestURL, incomingRequestURL)
			ctx = context.WithValue(ctx, CtxIncomingRequestMethod, c.Request().Method)

			c.SetRequest(c.Request().WithContext(ctx))

			return next(c)
		}
	}
}
