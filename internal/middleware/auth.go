package middleware

import (
	"context"

	"github.com/gate149/core/pkg"
	"github.com/gofiber/fiber/v2"
	ory "github.com/ory/client-go"
)

const sessionKey = "session"
const userIDKey = "user_id"

// OathkeeperMiddleware extracts user ID from X-User-ID header set by Oathkeeper
func OathkeeperMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Oathkeeper should set X-User-ID header after validating session
		userID := c.Get("X-User-ID")

		if userID != "" {
			c.Locals(userIDKey, userID)
			// Create a minimal session object for compatibility
			active := true
			session := &ory.Session{
				Identity: &ory.Identity{
					Id: userID,
				},
				Active: &active,
			}
			c.Locals(sessionKey, session)
		}

		return c.Next()
	}
}

func GetSession(ctx context.Context) (*ory.Session, error) {
	u, ok := ctx.Value(sessionKey).(*ory.Session)
	if !ok {
		return nil, pkg.Wrap(pkg.ErrUnauthenticated, nil, "", "no session in ctx")
	}
	return u, nil
}

// GetUserID extracts user ID from context
func GetUserID(ctx context.Context) (string, error) {
	if userID, ok := ctx.Value(userIDKey).(string); ok && userID != "" {
		return userID, nil
	}

	session, err := GetSession(ctx)
	if err != nil {
		return "", err
	}

	if session.Identity != nil && session.Identity.Id != "" {
		return session.Identity.Id, nil
	}

	return "", pkg.Wrap(pkg.ErrUnauthenticated, nil, "", "no user ID found")
}
