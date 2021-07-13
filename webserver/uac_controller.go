package webserver

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func UACRoutes(httpRouter *gin.Engine) {
	uacsGroup := httpRouter.Group("/uacs")
	{
		uacsGroup.GET("/generate/:instrumentName", UACGeneratorEndpoint)
	}
}

func UACGeneratorEndpoint(context *gin.Context) {
	context.JSON(http.StatusOK, nil)
}
