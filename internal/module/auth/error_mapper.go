package auth

import (
	"errors"

	"github.com/tranphuocnhan/radio-shuffle/internal/platform/response"

	"github.com/gin-gonic/gin"
)

func MapError(c *gin.Context, err error) bool {
	switch {
	case errors.Is(err, ErrEmailTaken):
		response.Conflict(c, err.Error())
		return true
	case errors.Is(err, ErrWeakPassword):
		response.BadRequest(c, err.Error())
		return true
	case errors.Is(err, ErrInvalidCredentials):
		response.Unauthorized(c, err.Error())
		return true
	case errors.Is(err, ErrInvalidRefreshToken):
		response.Unauthorized(c, err.Error())
		return true
	default:
		return false
	}
}
