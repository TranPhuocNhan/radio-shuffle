package syncer

import (
	"errors"

	"github.com/tranphuocnhan/radio-shuffle/internal/platform/response"

	"github.com/gin-gonic/gin"
)

func MapError(c *gin.Context, err error) bool {
	switch {
	case errors.Is(err, ErrJobNotFound):
		response.NotFound(c, "sync job not found")
		return true
	default:
		return false
	}
}
