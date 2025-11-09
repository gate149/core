package middleware

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/gate149/core/pkg"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type errorResponse struct {
	Err       string `json:"error"`
	Msg       string `json:"message"`
	RequestID string `json:"request_id,omitempty"`
}

// RequestLoggerMiddleware logs all incoming requests with timing and context
func RequestLoggerMiddleware(logger *slog.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Generate request ID if not present
		requestID := c.Get("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}
		c.Set("X-Request-ID", requestID)
		c.Locals("request_id", requestID)

		// Record start time
		start := time.Now()

		// Process request
		err := c.Next()

		// Calculate duration
		duration := time.Since(start)
		statusCode := c.Response().StatusCode()

		// Build log message
		logMsg := fmt.Sprintf("%s %s -> %d %s (%dms)",
			c.Method(),
			c.Path(),
			statusCode,
			http.StatusText(statusCode),
			duration.Milliseconds(),
		)

		// Build log attributes
		logAttrs := []any{
			slog.String("request_id", requestID),
			slog.String("method", c.Method()),
			slog.String("path", c.Path()),
			slog.Int("status", statusCode),
			slog.Int64("duration_ms", duration.Milliseconds()),
			slog.String("ip", c.IP()),
		}

		// Add user context if available
		if userID := c.Get("X-User-ID"); userID != "" {
			logAttrs = append(logAttrs, slog.String("user_id", userID))
		}
		if sessionID := c.Get("X-Session-ID"); sessionID != "" {
			logAttrs = append(logAttrs, slog.String("session_id", sessionID))
		}

		// Add query params if present
		if len(c.Queries()) > 0 {
			logAttrs = append(logAttrs, slog.Any("query", c.Queries()))
		}

		// Log based on status code (errors are already logged by ErrorHandlerMiddleware)
		if err == nil {
			if statusCode >= 500 {
				logger.Error(logMsg, logAttrs...)
			} else if statusCode >= 400 {
				logger.Warn(logMsg, logAttrs...)
			} else {
				logger.Info(logMsg, logAttrs...)
			}
		}

		return err
	}
}

// ErrorHandlerMiddleware handles errors, maps them to HTTP status codes and logs them
func ErrorHandlerMiddleware(logger *slog.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		err := c.Next()
		if err == nil {
			return nil
		}

		// Get request ID from context
		requestID := ""
		if rid, ok := c.Locals("request_id").(string); ok {
			requestID = rid
		}
		if requestID == "" {
			requestID = uuid.New().String()
			c.Set("X-Request-ID", requestID)
		}

		statusCode := c.Response().StatusCode()

		var cErr *pkg.CustomError
		if errors.As(err, &cErr) {
			statusCode = pkg.ToREST(err)
		}

		resp := errorResponse{
			Err:       http.StatusText(statusCode),
			Msg:       "",
			RequestID: requestID,
		}

		var fErr *fiber.Error
		if errors.As(err, &fErr) {
			statusCode = fErr.Code
			resp.Err = http.StatusText(statusCode)
			resp.Msg = fErr.Message
		}

		// Build log message
		logMsg := fmt.Sprintf("%s %s -> %d %s",
			c.Method(),
			c.Path(),
			statusCode,
			http.StatusText(statusCode),
		)

		// Build log attributes
		logAttrs := []any{
			slog.String("request_id", requestID),
			slog.String("method", c.Method()),
			slog.String("path", c.Path()),
			slog.Int("status", statusCode),
			slog.String("error", err.Error()),
			slog.String("ip", c.IP()),
		}

		// Add user context if available
		if userID := c.Get("X-User-ID"); userID != "" {
			logAttrs = append(logAttrs, slog.String("user_id", userID))
		}
		if sessionID := c.Get("X-Session-ID"); sessionID != "" {
			logAttrs = append(logAttrs, slog.String("session_id", sessionID))
		}

		if cErr != nil {
			resp.Msg = cErr.Message

			logAttrs = append(logAttrs,
				slog.String("operation", cErr.Op),
				slog.String("message", cErr.Message),
			)

			if cErr.Basic != nil {
				logAttrs = append(logAttrs, slog.String("error_type", fmt.Sprintf("%v", cErr.Basic)))
			}
			if cErr.Cause != nil {
				logAttrs = append(logAttrs, slog.String("cause", fmt.Sprintf("%v", cErr.Cause)))
			}
		}

		// Log error based on severity
		switch statusCode {
		case http.StatusInternalServerError:
			logger.Error(logMsg, logAttrs...)
		case http.StatusBadRequest:
			logger.Warn(logMsg, logAttrs...)
		case http.StatusUnauthorized, http.StatusForbidden:
			logger.Warn(logMsg, logAttrs...)
		case http.StatusNotFound:
			logger.Info(logMsg, logAttrs...)
		default:
			if statusCode >= 500 {
				logger.Error(logMsg, logAttrs...)
			} else if statusCode >= 400 {
				logger.Warn(logMsg, logAttrs...)
			}
		}

		return c.Status(statusCode).JSON(resp)
	}
}
