package mw

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log/slog"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	HeaderRequestID     = "X-Request-ID"
	ContextRequestIDKey = "request_id"
)

func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		rid := c.GetHeader(HeaderRequestID)
		if rid == "" {
			buf := make([]byte, 16)
			if _, err := rand.Read(buf); err == nil {
				rid = hex.EncodeToString(buf)
			} else {
				rid = "unknown"
			}
		}
		c.Writer.Header().Set(HeaderRequestID, rid)
		c.Set(ContextRequestIDKey, rid)
		c.Next()
	}
}

func Recover() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered any) {
		attrs := RequestLogAttrs(c,
			"panic_type", fmt.Sprintf("%T", recovered),
			"panic_message", fmt.Sprint(recovered),
			"stack", string(debug.Stack()),
		)
		slog.Error("panic", attrs...)
		c.AbortWithStatus(http.StatusInternalServerError)
	})
}

func LoggerStructured() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		attrs := RequestLogAttrs(c,
			"status", c.Writer.Status(),
			"latency_ms", time.Since(start).Milliseconds(),
			"bytes_out", c.Writer.Size(),
		)
		logRequest(c.Writer.Status(), attrs)
	}
}

func logRequest(status int, attrs []any) {
	switch {
	case status >= http.StatusInternalServerError:
		slog.Error("http_request", attrs...)
	case status >= http.StatusBadRequest:
		slog.Warn("http_request", attrs...)
	default:
		slog.Info("http_request", attrs...)
	}
}

func RequestIDFromContext(c *gin.Context) string {
	if value, ok := c.Get(ContextRequestIDKey); ok {
		if rid, ok := value.(string); ok {
			return rid
		}
	}
	return c.Writer.Header().Get(HeaderRequestID)
}

func RequestLogAttrs(c *gin.Context, extra ...any) []any {
	attrs := []any{
		"request_id", RequestIDFromContext(c),
		"method", c.Request.Method,
		"path", c.Request.URL.Path,
		"route", routePath(c),
		"query", c.Request.URL.RawQuery,
		"client_ip", c.ClientIP(),
		"user_agent", c.Request.UserAgent(),
	}
	if userID, ok := UserIDFromContext(c); ok {
		attrs = append(attrs, "user_id", userID)
	}
	return append(attrs, extra...)
}

func routePath(c *gin.Context) string {
	if route := c.FullPath(); route != "" {
		return route
	}
	return c.Request.URL.Path
}
