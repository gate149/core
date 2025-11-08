package middleware

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/gate149/core/pkg"
	"github.com/gofiber/fiber/v2"
)

type errorResponse struct {
	Err string `json:"error"`
	Msg string `json:"message"`
}

// ErrorHandlerMiddleware handles errors, maps them to HTTP status codes and logs them
func ErrorHandlerMiddleware(logger *slog.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		err := c.Next()
		if err == nil {
			return nil
		}

		statusCode := c.Response().StatusCode()

		var cErr *pkg.CustomError
		if errors.As(err, &cErr) {
			statusCode = pkg.ToREST(err)
		}

		resp := errorResponse{
			Err: http.StatusText(statusCode),
			Msg: "",
		}

		var fErr *fiber.Error
		if errors.As(err, &fErr) {
			statusCode = fErr.Code
			resp.Err = http.StatusText(statusCode)
			resp.Msg = fErr.Message
		}

		// Build log attributes
		logAttrs := []any{
			slog.String("method", c.Method()),
			slog.String("path", c.Path()),
			slog.Int("status", statusCode),
			slog.String("error", err.Error()),
		}

		if cErr != nil {
			resp.Msg = cErr.Message

			if cErr.Basic != nil {
				logAttrs = append(logAttrs, slog.Any("basic", cErr.Basic))
			}
			if cErr.Cause != nil {
				logAttrs = append(logAttrs, slog.Any("cause", cErr.Cause))
			}
			logAttrs = append(logAttrs,
				slog.String("operation", cErr.Op),
				slog.String("message", cErr.Message),
			)
		}

		switch statusCode {
		case http.StatusInternalServerError:
			logger.Error("Internal server error", logAttrs...)
		case http.StatusBadRequest:
			logger.Warn("Bad request", logAttrs...)
		case http.StatusUnauthorized, http.StatusForbidden:
			logger.Info("Authentication/Authorization error", logAttrs...)
		case http.StatusNotFound:
			logger.Info("Resource not found", logAttrs...)
		default:
			logger.Error("Unhandled error", logAttrs...)
		}

		return c.Status(statusCode).JSON(resp)
	}
}
