package mw

import (
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

const headerRequestID = "X-Request-ID"

func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		rid := c.GetHeader(headerRequestID)
		if rid == "" {
			buf := make([]byte, 16)
			if _, err := rand.Read(buf); err == nil {
				rid = hex.EncodeToString(buf)
			} else {
				rid = "unknown"
			}
		}
		c.Writer.Header().Set(headerRequestID, rid)
		c.Set("request_id", rid)
		c.Next()
	}
}

func Recover() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered any) {
		slog.Error("panic", "panic", recovered, "path", c.FullPath())
		c.AbortWithStatus(http.StatusInternalServerError)
	})
}

func LoggerStructured() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		slog.Info("http_request",
			"method", c.Request.Method,
			"path", c.FullPath(),
			"code", c.Writer.Status(),
			"ms", time.Since(start).Milliseconds(),
			headerRequestID, c.Writer.Header().Get(headerRequestID),
		)
	}
}
