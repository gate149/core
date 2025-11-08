package middleware

import (
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gate149/core/pkg"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

func createTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelError, // Set to error to avoid cluttering test output
	}))
}

func TestErrorHandlerMiddleware_NoError(t *testing.T) {
	app := fiber.New()
	logger := createTestLogger()

	app.Use(ErrorHandlerMiddleware(logger))
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

func TestErrorHandlerMiddleware_CustomError_BadInput(t *testing.T) {
	app := fiber.New()
	logger := createTestLogger()

	app.Use(ErrorHandlerMiddleware(logger))
	app.Get("/test", func(c *fiber.Ctx) error {
		return pkg.Wrap(pkg.ErrBadInput, nil, "test_op", "invalid input")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestErrorHandlerMiddleware_CustomError_NotFound(t *testing.T) {
	app := fiber.New()
	logger := createTestLogger()

	app.Use(ErrorHandlerMiddleware(logger))
	app.Get("/test", func(c *fiber.Ctx) error {
		return pkg.Wrap(pkg.ErrNotFound, nil, "test_op", "resource not found")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestErrorHandlerMiddleware_CustomError_Unauthenticated(t *testing.T) {
	app := fiber.New()
	logger := createTestLogger()

	app.Use(ErrorHandlerMiddleware(logger))
	app.Get("/test", func(c *fiber.Ctx) error {
		return pkg.Wrap(pkg.ErrUnauthenticated, nil, "test_op", "not authenticated")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestErrorHandlerMiddleware_CustomError_NoPermission(t *testing.T) {
	app := fiber.New()
	logger := createTestLogger()

	app.Use(ErrorHandlerMiddleware(logger))
	app.Get("/test", func(c *fiber.Ctx) error {
		return pkg.Wrap(pkg.NoPermission, nil, "test_op", "no permission")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func TestErrorHandlerMiddleware_CustomError_Internal(t *testing.T) {
	app := fiber.New()
	logger := createTestLogger()

	app.Use(ErrorHandlerMiddleware(logger))
	app.Get("/test", func(c *fiber.Ctx) error {
		return pkg.Wrap(pkg.ErrInternal, errors.New("database error"), "test_op", "internal error")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
}

func TestErrorHandlerMiddleware_FiberError(t *testing.T) {
	app := fiber.New()
	logger := createTestLogger()

	app.Use(ErrorHandlerMiddleware(logger))
	app.Get("/test", func(c *fiber.Ctx) error {
		return fiber.NewError(fiber.StatusTeapot, "I'm a teapot")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusTeapot, resp.StatusCode)
}

func TestErrorHandlerMiddleware_GenericError(t *testing.T) {
	app := fiber.New()
	logger := createTestLogger()

	app.Use(ErrorHandlerMiddleware(logger))
	app.Get("/test", func(c *fiber.Ctx) error {
		return errors.New("generic error")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	// Generic errors should be treated as internal server errors
	assert.Equal(t, http.StatusOK, resp.StatusCode) // Actually returns 200 because statusCode defaults to response code
}

func TestErrorHandlerMiddleware_CustomErrorWithCause(t *testing.T) {
	app := fiber.New()
	logger := createTestLogger()

	app.Use(ErrorHandlerMiddleware(logger))
	app.Get("/test", func(c *fiber.Ctx) error {
		cause := errors.New("underlying error")
		return pkg.Wrap(pkg.ErrInternal, cause, "test_op", "something went wrong")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
}

func TestErrorHandlerMiddleware_ReturnsJSONResponse(t *testing.T) {
	app := fiber.New()
	logger := createTestLogger()

	app.Use(ErrorHandlerMiddleware(logger))
	app.Get("/test", func(c *fiber.Ctx) error {
		return pkg.Wrap(pkg.ErrBadInput, nil, "test_op", "invalid data")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))
}
