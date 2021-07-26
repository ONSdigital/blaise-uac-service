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
		uacsGroup.GET("/instrument/:instrumentName/count", uacController.UACCountEndpoint)
		uacsGroup.POST("/uac", uacController.GetUacInfoEndpoint)
		uacsGroup.POST("/uac/attempts", uacController.IncrementPostcodeAttempts)
		uacsGroup.DELETE("/uac/attempts", uacController.ResetPostcodeAttempts)
		uacsGroup.DELETE("/admin/instrument/:instrumentName", uacController.AdminDeleteEndpoint)
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

func (UacController *UacController) UACCountEndpoint(context *gin.Context) {
	instrumentName := context.Param("instrumentName")

	uacCount, err := UacController.UacGenerator.GetUacCount(instrumentName)
	if err != nil {
		context.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	context.JSON(http.StatusOK, gin.H{"count": uacCount})
}

func (uacController *UacController) GetUacInfoEndpoint(context *gin.Context) {
	uac, err := uacController.getUacRequest(context)
	if err != nil {
		log.Println(err)
		context.AbortWithStatusJSON(http.StatusBadRequest, nil)
		return
	}

	uacInfo, err := uacController.UacGenerator.GetUacInfo(uac.UAC)
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

func (uacController *UacController) IncrementPostcodeAttempts(context *gin.Context) {
	uac, err := uacController.getUacRequest(context)
	if err != nil {
		log.Println(err)
		context.AbortWithStatusJSON(http.StatusBadRequest, nil)
		return
	}

	uacInfo, err := uacController.UacGenerator.IncrementPostcodeAttempts(uac.UAC)
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

func (uacController *UacController) ResetPostcodeAttempts(context *gin.Context) {
	uac, err := uacController.getUacRequest(context)
	if err != nil {
		log.Println(err)
		context.AbortWithStatusJSON(http.StatusBadRequest, nil)
		return
	}

	uacInfo, err := uacController.UacGenerator.ResetPostcodeAttempts(uac.UAC)
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

func (uacController *UacController) AdminDeleteEndpoint(context *gin.Context) {
	instrumentName := context.Param("instrumentName")
	err := uacController.UacGenerator.AdminDelete(instrumentName)
	if err != nil {
		log.Println(err)
		context.AbortWithStatusJSON(http.StatusInternalServerError, nil)
		return
	}
	context.JSON(http.StatusNoContent, nil)
}

func (uacController *UacController) blaiseRestApiError(context *gin.Context, err error) {
	if err.Error() == "Instrument not found" {
		context.AbortWithStatusJSON(http.StatusBadRequest, ResponseError{Error: err.Error()})
		return
	}
	context.AbortWithError(http.StatusInternalServerError, err)
}

func (uacController *UacController) getUacRequest(context *gin.Context) (UACRequest, error) {
	body, err := ioutil.ReadAll(context.Request.Body)
	if err != nil {
		return UACRequest{}, err
	}
	defer context.Request.Body.Close()

	var uac UACRequest
	err = json.Unmarshal(body, &uac)
	if err != nil {
		return UACRequest{}, err
	}
	return uac, nil
}
