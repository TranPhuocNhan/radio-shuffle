package mw

import (
	"strconv"
	"strings"

	"github.com/tranphuocnhan/radio-shuffle/internal/platform/response"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

const (
	ContextUserIDKey  = "user_id"
	ContextUserRoleKey = "user_role"
)

type accessClaims struct {
	jwt.RegisteredClaims
	Role string `json:"role"`
}

// AuthRequired validates JWT access tokens and injects user_id into context.
func AuthRequired(signingKey []byte, issuer string) gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := c.GetHeader("Authorization")
		if !strings.HasPrefix(auth, "Bearer ") {
			response.Unauthorized(c, "missing token")
			c.Abort()
			return
		}
		tokenStr := strings.TrimPrefix(auth, "Bearer ")

		claims := accessClaims{}
		token, err := jwt.ParseWithClaims(tokenStr, &claims, func(t *jwt.Token) (any, error) {
			return signingKey, nil
		})
		if err != nil || !token.Valid {
			response.Unauthorized(c, "invalid token")
			c.Abort()
			return
		}
		if claims.Issuer != issuer {
			response.Unauthorized(c, "invalid token")
			c.Abort()
			return
		}
		userID, err := strconv.ParseInt(claims.Subject, 10, 64)
		if err != nil {
			response.Unauthorized(c, "invalid subject")
			c.Abort()
			return
		}
		c.Set(ContextUserIDKey, userID)
		c.Set(ContextUserRoleKey, claims.Role)
		c.Next()
	}
}

// UserIDFromContext returns the user_id if it exists in context.
func UserIDFromContext(c *gin.Context) (int64, bool) {
	val, ok := c.Get(ContextUserIDKey)
	if !ok {
		return 0, false
	}
	id, ok := val.(int64)
	return id, ok
}

// UserRoleFromContext returns the user role if it exists in context.
func UserRoleFromContext(c *gin.Context) (string, bool) {
	val, ok := c.Get(ContextUserRoleKey)
	if !ok {
		return "", false
	}
	role, ok := val.(string)
	return role, ok
}

// AdminRequired enforces that the user role is admin.
func AdminRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, ok := UserRoleFromContext(c)
		if !ok || role != "admin" {
			response.Forbidden(c, "admin role required")
			c.Abort()
			return
		}
		c.Next()
	}
}

