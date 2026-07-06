package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/omnichannel/omni-chan-v2/services/storage-service/pkg/jwt"
	"github.com/omnichannel/omni-chan-v2/services/storage-service/pkg/response"
)

const claimsKey = "auth_claims"

type Actor struct {
	UserID uuid.UUID
	TeamID uuid.UUID
	Role   string
}

func Auth(jwtManager *jwt.Manager, internalToken string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if token := strings.TrimSpace(c.Get("X-Internal-Token")); token != "" && token == internalToken {
			teamRaw := strings.TrimSpace(c.Get("X-Team-ID"))
			userRaw := strings.TrimSpace(c.Get("X-User-ID"))
			teamID, err := uuid.Parse(teamRaw)
			if err != nil {
				return response.Unauthorized(c, "invalid X-Team-ID")
			}
			userID := uuid.Nil
			if userRaw != "" {
				userID, err = uuid.Parse(userRaw)
				if err != nil {
					return response.Unauthorized(c, "invalid X-User-ID")
				}
			}
			c.Locals(claimsKey, &jwt.Claims{UserID: userID, TeamID: teamID, Role: strings.TrimSpace(c.Get("X-User-Role"))})
			return c.Next()
		}

		header := c.Get("Authorization")
		if header == "" || !strings.HasPrefix(header, "Bearer ") {
			return response.Unauthorized(c, "missing authorization header")
		}
		token := strings.TrimSpace(strings.TrimPrefix(header, "Bearer "))
		claims, err := jwtManager.ParseAccess(token)
		if err != nil {
			return response.Unauthorized(c, "invalid token")
		}
		c.Locals(claimsKey, claims)
		return c.Next()
	}
}

func ActorFromContext(c *fiber.Ctx) (Actor, error) {
	value := c.Locals(claimsKey)
	claims, ok := value.(*jwt.Claims)
	if !ok || claims == nil {
		return Actor{}, fiber.ErrUnauthorized
	}
	return Actor{UserID: claims.UserID, TeamID: claims.TeamID, Role: claims.Role}, nil
}
