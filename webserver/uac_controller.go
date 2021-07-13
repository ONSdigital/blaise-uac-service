package webserver

import (
	"fmt"
	"net/http"

	"github.com/ONSDigital/blaise-uac-service/blaiserestapi"
	"github.com/ONSDigital/blaise-uac-service/uacgenerator"
	"github.com/gin-gonic/gin"
)

type ResponseError struct {
	Error string `json:"error"`
}

type UacController struct {
	BlaiseRestApi blaiserestapi.BlaiseRestApiInterface
	UacGenerator  uacgenerator.UacGeneratorInterface
}

func (uacController *UacController) AddRoutes(httpRouter *gin.Engine) {
	uacsGroup := httpRouter.Group("/uacs")
	{
		uacsGroup.POST("/:instrumentName", uacController.UACGeneratorEndpoint)
		// uacsGroup.GET("/:instrumentName", uacController.UACGetEndpoint)
	}
}

func (uacController *UacController) UACGeneratorEndpoint(context *gin.Context) {
	instrumentName := context.Param("instrumentName")
	instrumentModes, err := uacController.BlaiseRestApi.GetInstrumentModes(instrumentName)
	if err != nil {
		uacController.blaiseRestApiError(context, err)
		return
	}
	if !instrumentModes.HasCawi() {
		context.AbortWithStatusJSON(http.StatusBadRequest, ResponseError{Error: fmt.Sprintf("Instrument '%s' is not installed in CAWI mode", instrumentName)})
		return
	}
	caseIDs, err := uacController.BlaiseRestApi.GetCaseIds(instrumentName)
	if err != nil {
		uacController.blaiseRestApiError(context, err)
		return
	}
	err = uacController.UacGenerator.Generate(instrumentName, caseIDs)
	if err != nil {
		context.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	context.JSON(http.StatusOK, nil)
}

func (uacController *UacController) blaiseRestApiError(context *gin.Context, err error) {
	if err.Error() == "Instrument not found" {
		context.AbortWithStatusJSON(http.StatusBadRequest, ResponseError{Error: err.Error()})
		return
	}
	context.AbortWithError(http.StatusInternalServerError, err)
}
