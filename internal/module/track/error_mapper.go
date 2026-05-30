package track

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
		response.NotFound(c, "track not found")
		return true
	default:
		return false
	}
}
