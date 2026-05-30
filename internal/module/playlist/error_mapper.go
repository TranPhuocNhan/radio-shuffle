package playlist

import (
	"errors"

	"github.com/tranphuocnhan/radio-shuffle/internal/platform/response"

	"github.com/gin-gonic/gin"
)

func MapError(c *gin.Context, err error) bool {
	switch {
	case errors.Is(err, ErrNotFound):
		response.NotFound(c, "playlist not found")
		return true
	case errors.Is(err, ErrForbidden):
		response.Forbidden(c, "forbidden")
		return true
	case errors.Is(err, ErrTrackNotFound):
		response.NotFound(c, "track not found")
		return true
	case errors.Is(err, ErrDuplicateTrack):
		response.Conflict(c, "duplicate track")
		return true
	case errors.Is(err, ErrDuplicatePosition):
		response.Conflict(c, "duplicate position")
		return true
	case errors.Is(err, ErrInvalidPosition):
		response.BadRequest(c, "invalid position")
		return true
	case errors.Is(err, ErrInvalidTrackOrder):
		response.BadRequest(c, "invalid track reorder payload")
		return true
	default:
		return false
	}
}
