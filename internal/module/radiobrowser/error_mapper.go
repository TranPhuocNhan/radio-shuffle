package radiobrowser

import "github.com/gin-gonic/gin"

func MapError(_ *gin.Context, _ error) bool {
	return false
}
