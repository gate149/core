package middleware

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	ory "github.com/ory/client-go"
	"github.com/stretchr/testify/assert"
)

func TestOathkeeperMiddleware_WithUserID(t *testing.T) {
	app := fiber.New()

	var extractedUserID string
	var sessionExists bool

	app.Use(OathkeeperMiddleware())
	app.Get("/test", func(c *fiber.Ctx) error {
		extractedUserID = c.Locals(userIDKey).(string)
		sessionExists = c.Locals(sessionKey) != nil
		return c.SendString("ok")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-User-ID", "test-user-id")
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, "test-user-id", extractedUserID)
	assert.True(t, sessionExists)
}

func TestOathkeeperMiddleware_WithoutUserID(t *testing.T) {
	app := fiber.New()

	var userIDExists bool

	app.Use(OathkeeperMiddleware())
	app.Get("/test", func(c *fiber.Ctx) error {
		userIDExists = c.Locals(userIDKey) != nil
		return c.SendString("ok")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	assert.False(t, userIDExists)
}

func TestOathkeeperMiddleware_EmptyUserID(t *testing.T) {
	app := fiber.New()

	var userIDExists bool

	app.Use(OathkeeperMiddleware())
	app.Get("/test", func(c *fiber.Ctx) error {
		userIDExists = c.Locals(userIDKey) != nil
		return c.SendString("ok")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-User-ID", "")
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	assert.False(t, userIDExists)
}

func TestOathkeeperMiddleware_CreatesValidSession(t *testing.T) {
	app := fiber.New()

	var sessionActive bool
	var sessionIdentityID string

	app.Use(OathkeeperMiddleware())
	app.Get("/test", func(c *fiber.Ctx) error {
		if session := c.Locals(sessionKey); session != nil {
			if s, ok := session.(*ory.Session); ok {
				if s.Active != nil {
					sessionActive = *s.Active
				}
				if s.Identity != nil {
					sessionIdentityID = s.Identity.Id
				}
			}
		}
		return c.SendString("ok")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-User-ID", "test-user-123")
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	assert.True(t, sessionActive)
	assert.Equal(t, "test-user-123", sessionIdentityID)
}

func TestGetUserID_WithUserIDInLocals(t *testing.T) {
	app := fiber.New()

	var retrievedUserID string
	var getUserIDErr error

	app.Use(OathkeeperMiddleware())
	app.Get("/test", func(c *fiber.Ctx) error {
		retrievedUserID, getUserIDErr = GetUserID(c.Context())
		return c.SendString("ok")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-User-ID", "test-user-456")
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	assert.NoError(t, getUserIDErr)
	assert.Equal(t, "test-user-456", retrievedUserID)
}

func TestGetUserID_NoUserID(t *testing.T) {
	app := fiber.New()

	var getUserIDErr error

	app.Get("/test", func(c *fiber.Ctx) error {
		_, getUserIDErr = GetUserID(c.Context())
		return c.SendString("ok")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Error(t, getUserIDErr)
}

func TestGetSession_WithSession(t *testing.T) {
	app := fiber.New()

	var getSessionErr error
	var sessionExists bool

	app.Use(OathkeeperMiddleware())
	app.Get("/test", func(c *fiber.Ctx) error {
		session, err := GetSession(c.Context())
		getSessionErr = err
		sessionExists = session != nil
		return c.SendString("ok")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-User-ID", "test-user-789")
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	assert.NoError(t, getSessionErr)
	assert.True(t, sessionExists)
}

func TestGetSession_NoSession(t *testing.T) {
	app := fiber.New()

	var getSessionErr error

	app.Get("/test", func(c *fiber.Ctx) error {
		_, err := GetSession(c.Context())
		getSessionErr = err
		return c.SendString("ok")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Error(t, getSessionErr)
}
