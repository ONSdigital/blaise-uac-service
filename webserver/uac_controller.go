package webserver

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"cloud.google.com/go/datastore"
	"github.com/ONSDigital/blaise-uac-service/blaiserestapi"
	"github.com/ONSDigital/blaise-uac-service/uacgenerator"
	"github.com/gin-gonic/gin"
)

type ResponseError struct {
	Error string `json:"error"`
}

type UACRequest struct {
	UAC string `json:"uac"`
}

type UacController struct {
	BlaiseRestApi blaiserestapi.BlaiseRestApiInterface
	UacGenerator  uacgenerator.UacGeneratorInterface
}

func (uacController *UacController) AddRoutes(httpRouter *gin.Engine) {
	uacsGroup := httpRouter.Group("/uacs")
	{
		uacsGroup.POST("/instrument/:instrumentName", uacController.UACGeneratorEndpoint)
		uacsGroup.GET("/instrument/:instrumentName", uacController.UACGetAllEndpoint)
		uacsGroup.POST("/uac", uacController.GetUacInfoEndpoint)
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
	uacs, err := uacController.UacGenerator.GetAllUacs(instrumentName)
	if err != nil {
		context.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	context.JSON(http.StatusOK, uacs)
}

func (UacController *UacController) UACGetAllEndpoint(context *gin.Context) {
	instrumentName := context.Param("instrumentName")

	uacs, err := UacController.UacGenerator.GetAllUacs(instrumentName)
	if err != nil {
		context.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	context.JSON(http.StatusOK, uacs)
}

func (UacController *UacController) GetUacInfoEndpoint(context *gin.Context) {
	body, err := ioutil.ReadAll(context.Request.Body)
	if err != nil {
		log.Println(err)
		context.AbortWithStatusJSON(http.StatusBadRequest, nil)
		return
	}
	defer context.Request.Body.Close()

	var uac UACRequest
	err = json.Unmarshal(body, &uac)
	if err != nil {
		log.Println(err)
		context.AbortWithStatusJSON(http.StatusBadRequest, nil)
		return
	}

	uacInfo, err := UacController.UacGenerator.GetUacInfo(uac.UAC)
	if err != nil {
		if err == datastore.ErrNoSuchEntity {
			context.JSON(http.StatusNotFound, nil)
			return
		}
		log.Println(err)
		context.AbortWithStatusJSON(http.StatusInternalServerError, nil)
		return
	}
	context.JSON(http.StatusOK, uacInfo)
}

func (uacController *UacController) blaiseRestApiError(context *gin.Context, err error) {
	if err.Error() == "Instrument not found" {
		context.AbortWithStatusJSON(http.StatusBadRequest, ResponseError{Error: err.Error()})
		return
	}
	context.AbortWithError(http.StatusInternalServerError, err)
}
