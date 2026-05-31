package apidocs

import (
	"net/http"

	projectdocs "github.com/tranphuocnhan/radio-shuffle/docs"

	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func RegisterRoutes(r *gin.Engine) {
	r.GET("/swagger", SwaggerUI)
	r.GET("/swagger/*any", ginSwagger.WrapHandler(
		swaggerfiles.Handler,
		ginSwagger.URL("/openapi.yaml"),
		ginSwagger.DefaultModelsExpandDepth(-1),
	))
	r.GET("/openapi.yaml", OpenAPIYAML)
}

func SwaggerUI(c *gin.Context) {
	c.Redirect(http.StatusFound, "/swagger/index.html")
}

func OpenAPIYAML(c *gin.Context) {
	c.Data(http.StatusOK, "application/yaml; charset=utf-8", projectdocs.OpenAPIYAML)
}
