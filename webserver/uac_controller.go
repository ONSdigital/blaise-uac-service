package webserver

import (
	"net/http"

	"github.com/ONSDigital/blaise-uac-service/blaiserestapi"
	"github.com/ONSDigital/blaise-uac-service/uacgenerator"
	"github.com/gin-gonic/gin"
)

type UacController struct {
	BlaiseRestApi *blaiserestapi.BlaiseRestApi
	UacGenerator  *uacgenerator.UacGenerator
}

func (uacController *UacController) AddRoutes(httpRouter *gin.Engine) {
	uacsGroup := httpRouter.Group("/uacs")
	{
		uacsGroup.GET("/generate/:instrumentName", UACGeneratorEndpoint)
	}
}

func UACGeneratorEndpoint(context *gin.Context) {
	context.JSON(http.StatusOK, nil)
}
