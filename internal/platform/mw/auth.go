package mw

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

const ContextUserIDKey = "user_id"

// AuthRequired validates JWT access tokens and injects user_id into context.
func AuthRequired(signingKey []byte, issuer string) gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := c.GetHeader("Authorization")
		if !strings.HasPrefix(auth, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
			return
		}
		tokenStr := strings.TrimPrefix(auth, "Bearer ")

		claims := jwt.RegisteredClaims{}
		token, err := jwt.ParseWithClaims(tokenStr, &claims, func(t *jwt.Token) (any, error) {
			return signingKey, nil
		})
		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}
		if claims.Issuer != issuer {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}
		userID, err := strconv.ParseInt(claims.Subject, 10, 64)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid subject"})
			return
		}
		c.Set(ContextUserIDKey, userID)
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
