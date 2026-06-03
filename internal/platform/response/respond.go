package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Reply matches the project's API envelope (`success`, `data`, `error`, optional `meta`).
func ReplyJSON(c *gin.Context, httpStatus int, body map[string]any) {
	c.JSON(httpStatus, body)
}

// OK sends `{ "success": true, "data": data, "error": null }`.
func OK(c *gin.Context, httpStatus int, data any) {
	ReplyJSON(c, httpStatus, map[string]any{
		"success": true,
		"data":    data,
		"error":   nil,
	})
}

// Paginated wraps list payloads with paging metadata ({ page, limit, total }).
func Paginated(c *gin.Context, httpStatus int, data any, page, limit int, total int64) {
	meta := map[string]any{"page": page, "limit": limit, "total": total}
	ReplyJSON(c, httpStatus, map[string]any{
		"success": true,
		"data":    data,
		"error":   nil,
		"meta":    meta,
	})
}

// ErrBody mirrors the error object in failing responses.
type ErrBody struct {
	Code    string         `json:"code"`
	Message string         `json:"message"`
	Details map[string]any `json:"details,omitempty"`
}

// Error replies with `{ success: false, data: null, error: {...} }`.
func Error(c *gin.Context, httpStatus int, eb ErrBody) {
	ReplyJSON(c, httpStatus, map[string]any{
		"success": false,
		"data":    nil,
		"error": map[string]any{
			"code":    eb.Code,
			"message": eb.Message,
			"details": eb.Details,
		},
	})
}

func BadRequest(c *gin.Context, message string) {
	Error(c, http.StatusBadRequest, ErrBody{
		Code:    "VALIDATION_ERROR",
		Message: message,
		Details: nil,
	})
}

func Unauthorized(c *gin.Context, message string) {
	Error(c, http.StatusUnauthorized, ErrBody{Code: "UNAUTHORIZED", Message: message})
}

func Forbidden(c *gin.Context, message string) {
	Error(c, http.StatusForbidden, ErrBody{Code: "FORBIDDEN", Message: message})
}

func NotFound(c *gin.Context, message string) {
	Error(c, http.StatusNotFound, ErrBody{Code: "NOT_FOUND", Message: message})
}

func Conflict(c *gin.Context, message string) {
	Error(c, http.StatusConflict, ErrBody{Code: "CONFLICT", Message: message})
}

func Internal(c *gin.Context, _ error) {
	Error(c, http.StatusInternalServerError, ErrBody{
		Code:    "INTERNAL_ERROR",
		Message: "internal error",
		Details: nil,
	})
}
