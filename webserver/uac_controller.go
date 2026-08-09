package webserver

import (
	"encoding/json"
	"fmt"
	"io"
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

type UacRequest struct {
	Uac string `json:"uac"`
}

type UacGenerateRequest struct {
	InstrumentName string   `json:"instrument_name"`
	CaseIDs        []string `json:"case_ids"`
}

type UacController struct {
	BlaiseRestApi blaiserestapi.BlaiseRestApiInterface
	UacService  uacgenerator.UacServiceInterface
}

func (uacController *UacController) AddRoutes(httpRouter *gin.Engine) {
	uacsGroup := httpRouter.Group("/uacs")
	{
		uacsGroup.POST("/instrument/:instrumentName", uacController.UacInstrumentGenerateEndpoint)
		uacsGroup.GET("/instrument/:instrumentName", uacController.UacGetAllEndpoint)
		uacsGroup.GET("/instrument/:instrumentName/bycaseid", uacController.UacGetAllByCaseIDEndpoint)
		uacsGroup.GET("/instrument/:instrumentName/count", uacController.UacCountEndpoint)
		uacsGroup.POST("/generate", uacController.UacGenerateEndpoint)
		uacsGroup.POST("/uac", uacController.GetUacInfoEndpoint)
		uacsGroup.DELETE("/admin/instrument/:instrumentName", uacController.AdminDeleteEndpoint)
		uacsGroup.GET("/instruments", uacController.ListInstrumentsEndpoint)
		uacsGroup.POST("/import", uacController.ImportEndpoint)
		uacsGroup.PATCH("/uac/disable/:uac", uacController.UacDisableEndpoint)
		uacsGroup.PATCH("/uac/enable/:uac", uacController.UacEnableEndpoint)
		uacsGroup.GET("/uac/:instrumentName/disabled", uacController.UacGetAllDisabledEndpoint)

	}
}

func (uacController *UacController) UacInstrumentGenerateEndpoint(context *gin.Context) {
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
	err = uacController.UacService.Generate(instrumentName, caseIDs)
	if err != nil {
		_ = context.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	uacs, err := uacController.UacService.GetAllUacs(instrumentName)
	if err != nil {
		_ = context.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	uacs.BuildUacChunks()
	context.JSON(http.StatusOK, uacs)
}

func (uacController *UacController) UacGenerateEndpoint(context *gin.Context) {
	body, err := io.ReadAll(context.Request.Body)
	if err != nil {
		_ = context.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	defer context.Request.Body.Close()
	var uacGenerateRequest UacGenerateRequest
	err = json.Unmarshal(body, &uacGenerateRequest)
	if err != nil {
		_ = context.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	if uacGenerateRequest.InstrumentName == "" {
		context.AbortWithStatusJSON(http.StatusBadRequest, ResponseError{Error: "Must provide instrument name"})
		return
	}
	err = uacController.UacService.Generate(uacGenerateRequest.InstrumentName, uacGenerateRequest.CaseIDs)
	if err != nil {
		_ = context.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	uacs, err := uacController.UacService.GetAllUacs(uacGenerateRequest.InstrumentName)
	if err != nil {
		_ = context.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	uacs.BuildUacChunks()
	context.JSON(http.StatusOK, uacs)
}

func (uacController *UacController) UacGetAllEndpoint(context *gin.Context) {
	instrumentName := context.Param("instrumentName")

	uacs, err := uacController.UacService.GetAllUacs(instrumentName)
	if err != nil {
		_ = context.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	uacs.BuildUacChunks()
	context.JSON(http.StatusOK, uacs)
}

func (uacController *UacController) UacGetAllByCaseIDEndpoint(context *gin.Context) {
	instrumentName := context.Param("instrumentName")

	uacs, err := uacController.UacService.GetAllUacsByCaseID(instrumentName)
	if err != nil {
		_ = context.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	uacs.BuildUacChunks()
	context.JSON(http.StatusOK, uacs)
}

func (uacController *UacController) ListInstrumentsEndpoint(context *gin.Context) {
	instrumentNames, err := uacController.UacService.GetInstruments()
	if err != nil {
		_ = context.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	context.JSON(http.StatusOK, instrumentNames)
}

func (uacController *UacController) UacCountEndpoint(context *gin.Context) {
	instrumentName := context.Param("instrumentName")

	uacCount, err := uacController.UacService.GetUacCount(instrumentName)
	if err != nil {
		_ = context.AbortWithError(http.StatusInternalServerError, err)
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

	uacInfo, err := uacController.UacService.GetUacInfo(uac.Uac)
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
	err := uacController.UacService.AdminDelete(instrumentName)
	if err != nil {
		log.Println(err)
		context.AbortWithStatusJSON(http.StatusInternalServerError, nil)
		return
	}
	context.JSON(http.StatusNoContent, nil)
}

func (uacController *UacController) ImportEndpoint(context *gin.Context) {
	body, err := io.ReadAll(context.Request.Body)
	if err != nil {
		_ = context.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	defer context.Request.Body.Close()
	var uacs []string
	err = json.Unmarshal(body, &uacs)
	if err != nil {
		_ = context.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	importCount, err := uacController.UacService.ImportUacs(uacs)
	if err != nil {
		if _, ok := err.(*uacgenerator.ImportError); ok {
			context.AbortWithStatusJSON(http.StatusBadRequest, ResponseError{Error: err.Error()})
			return
		}
		_ = context.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	context.JSON(http.StatusOK, gin.H{"uacs_imported": importCount})
}

func (uacController *UacController) blaiseRestApiError(context *gin.Context, err error) {
	if err.Error() == "Instrument not found" {
		context.AbortWithStatusJSON(http.StatusBadRequest, ResponseError{Error: err.Error()})
		return
	}
	if err.Error() == "invalid uac" {
		context.AbortWithStatusJSON(http.StatusBadRequest, ResponseError{Error: err.Error()})
		return
	}
	_ = context.AbortWithError(http.StatusInternalServerError, err)
}

func (uacController *UacController) getUacRequest(context *gin.Context) (UacRequest, error) {
	body, err := io.ReadAll(context.Request.Body)
	if err != nil {
		return UacRequest{}, err
	}
	defer context.Request.Body.Close()

	var uac UacRequest
	err = json.Unmarshal(body, &uac)
	if err != nil {
		return UacRequest{}, err
	}
	return uac, nil
}

func (uacController *UacController) UacDisableEndpoint(context *gin.Context) {
	uac := context.Param("uac")

	err := uacController.UacService.DisableUac(uac)
	if err != nil {
		uacController.blaiseRestApiError(context, err)
		return
	}
	context.JSON(http.StatusOK, nil)
}

func (uacController *UacController) UacEnableEndpoint(context *gin.Context) {
	uac := context.Param("uac")

	err := uacController.UacService.EnableUac(uac)
	if err != nil {
		uacController.blaiseRestApiError(context, err)
		return
	}
	context.JSON(http.StatusOK, nil)
}

func (uacController *UacController) UacGetAllDisabledEndpoint(context *gin.Context) {
	instrumentName := context.Param("instrumentName")

	uacs, err := uacController.UacService.GetAllUacsDisabled(instrumentName)
	if err != nil {
		uacController.blaiseRestApiError(context, err)
		return
	}
	uacs.BuildUacChunks()
	context.JSON(http.StatusOK, uacs)
}
