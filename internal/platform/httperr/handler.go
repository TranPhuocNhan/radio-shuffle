package httperr

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/tranphuocnhan/radio-shuffle/internal/platform/mw"
	"github.com/tranphuocnhan/radio-shuffle/internal/platform/response"

	"github.com/gin-gonic/gin"
)

type AppHandler func(c *gin.Context) error

type Mapper func(c *gin.Context, err error) bool

type HTTPError struct {
	Status  int
	Message string
}

func (e *HTTPError) Error() string {
	return e.Message
}

func New(status int, message string) error {
	return &HTTPError{Status: status, Message: message}
}

func BadRequest(message string) error {
	return New(http.StatusBadRequest, message)
}

func Unauthorized(message string) error {
	return New(http.StatusUnauthorized, message)
}

func Wrap(fn AppHandler, mappers ...Mapper) gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := fn(c); err != nil {
			Handle(c, err, mappers...)
		}
	}
}

func Handle(c *gin.Context, err error, mappers ...Mapper) {
	if err == nil {
		return
	}

	for _, mapper := range mappers {
		if mapper != nil && mapper(c, err) {
			logHTTPError(c, err, c.Writer.Status())
			return
		}
	}

	var httpErr *HTTPError
	if errors.As(err, &httpErr) {
		writeHTTPError(c, httpErr)
		logHTTPError(c, err, httpErr.Status)
		return
	}

	response.Internal(c, err)
	logHTTPError(c, err, http.StatusInternalServerError)
}

func writeHTTPError(c *gin.Context, httpErr *HTTPError) {
	switch httpErr.Status {
	case http.StatusBadRequest:
		response.BadRequest(c, httpErr.Message)
	case http.StatusUnauthorized:
		response.Unauthorized(c, httpErr.Message)
	case http.StatusForbidden:
		response.Forbidden(c, httpErr.Message)
	case http.StatusNotFound:
		response.NotFound(c, httpErr.Message)
	case http.StatusConflict:
		response.Conflict(c, httpErr.Message)
	default:
		response.Error(c, httpErr.Status, response.ErrBody{
			Code:    http.StatusText(httpErr.Status),
			Message: httpErr.Message,
		})
	}
}

func logHTTPError(c *gin.Context, err error, status int) {
	attrs := mw.RequestLogAttrs(c,
		"status", status,
		"err", err,
		"error_type", errorType(err),
	)
	if status >= http.StatusInternalServerError {
		slog.Error("http_error", attrs...)
		return
	}
	slog.Warn("http_error", attrs...)
}

func errorType(err error) string {
	if err == nil {
		return ""
	}
	return fmt.Sprintf("%T", err)
}
