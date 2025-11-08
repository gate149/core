package health

import (
	"github.com/gofiber/fiber/v2"
)

type HealthHandlers struct{}

func NewHandlers() *HealthHandlers {
	return &HealthHandlers{}
}

func (h *HealthHandlers) GetHealth(c *fiber.Ctx) error {
	return c.JSON(map[string]string{
		"status":  "ok",
		"message": "Backend is running",
	})
}
