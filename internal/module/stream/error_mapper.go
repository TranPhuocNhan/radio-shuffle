package stream

import (
	"errors"

	"github.com/tranphuocnhan/radio-shuffle/internal/platform/response"

	"github.com/gin-gonic/gin"
)

func MapError(c *gin.Context, err error) bool {
	switch {
	case errors.Is(err, ErrStationNotFound):
		response.NotFound(c, "station not found")
		return true
	case errors.Is(err, ErrNotFound):
		response.NotFound(c, "stream not found")
		return true
	case errors.Is(err, ErrForbidden):
		response.Forbidden(c, "forbidden")
		return true
	case errors.Is(err, ErrAlreadyEnded):
		response.Conflict(c, "stream already ended")
		return true
	default:
		return false
	}
}
