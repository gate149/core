package health

import (
	"encoding/json"
	"io"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

func TestGetHealth_Success(t *testing.T) {
	app := fiber.New()
	handlers := NewHandlers()

	app.Get("/health", handlers.GetHealth)

	req := httptest.NewRequest("GET", "/health", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var response map[string]string
	body, _ := io.ReadAll(resp.Body)
	json.Unmarshal(body, &response)

	assert.Equal(t, "ok", response["status"])
	assert.Equal(t, "Backend is running", response["message"])
}

func TestGetHealth_ReturnsJSON(t *testing.T) {
	app := fiber.New()
	handlers := NewHandlers()

	app.Get("/health", handlers.GetHealth)

	req := httptest.NewRequest("GET", "/health", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))
}
