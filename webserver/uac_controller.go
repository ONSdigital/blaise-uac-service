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

type UACRequest struct {
	UAC string `json:"uac"`
}

type UACGenerateRequest struct {
	InstrumentName string   `json:"instrument_name"`
	CaseIDs        []string `json:"case_ids"`
	Disabled       bool     `json:"disabled" datastore:"disabled"`
}

type UacController struct {
	BlaiseRestApi blaiserestapi.BlaiseRestApiInterface
	UacGenerator  uacgenerator.UacGeneratorInterface
}

func (uacController *UacController) AddRoutes(httpRouter *gin.Engine) {
	uacsGroup := httpRouter.Group("/uacs")
	{
		uacsGroup.POST("/instrument/:instrumentName", uacController.UACInstrumentGenerateEndpoint)
		uacsGroup.GET("/instrument/:instrumentName", uacController.UACGetAllEndpoint)
		uacsGroup.GET("/instrument/:instrumentName/bycaseid", uacController.UACGetAllByCaseIDEndpoint)
		uacsGroup.GET("/instrument/:instrumentName/count", uacController.UACCountEndpoint)
		uacsGroup.POST("/generate", uacController.UACGenerateEndpoint)
		uacsGroup.POST("/uac", uacController.GetUacInfoEndpoint)
		uacsGroup.DELETE("/admin/instrument/:instrumentName", uacController.AdminDeleteEndpoint)
		uacsGroup.GET("/instruments", uacController.ListInstrumentsEndpoint)
		uacsGroup.POST("/import", uacController.ImportEndpoint)

		//uacsGroup.GET("/uac/:uac/disable", uacController.UACDisableEndpoint)
		//uacsGroup.GET("/uac/:uac/enable", uacController.UACEnableEndpoint)
		uacsGroup.GET("/uac/:instrumentName/disabled", uacController.UACGetAllDisabledEndpoint)

	}
}

func (uacController *UacController) UACInstrumentGenerateEndpoint(context *gin.Context) {
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
		_ = context.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	uacs, err := uacController.UacGenerator.GetAllUacs(instrumentName)
	if err != nil {
		_ = context.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	uacs.BuildUacChunks()
	context.JSON(http.StatusOK, uacs)
}

func (uacController *UacController) UACGenerateEndpoint(context *gin.Context) {
	body, err := io.ReadAll(context.Request.Body)
	if err != nil {
		_ = context.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	defer context.Request.Body.Close()
	var uacGenerateRequest UACGenerateRequest
	err = json.Unmarshal(body, &uacGenerateRequest)
	if err != nil {
		_ = context.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	if uacGenerateRequest.InstrumentName == "" {
		context.AbortWithStatusJSON(http.StatusBadRequest, ResponseError{Error: "Must provide instrument name"})
		return
	}
	err = uacController.UacGenerator.Generate(uacGenerateRequest.InstrumentName, uacGenerateRequest.CaseIDs)
	if err != nil {
		_ = context.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	uacs, err := uacController.UacGenerator.GetAllUacs(uacGenerateRequest.InstrumentName)
	if err != nil {
		_ = context.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	uacs.BuildUacChunks()
	context.JSON(http.StatusOK, uacs)
}

func (uacController *UacController) UACGetAllEndpoint(context *gin.Context) {
	instrumentName := context.Param("instrumentName")

	uacs, err := uacController.UacGenerator.GetAllUacs(instrumentName)
	if err != nil {
		_ = context.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	uacs.BuildUacChunks()
	context.JSON(http.StatusOK, uacs)
}

func (uacController *UacController) UACGetAllByCaseIDEndpoint(context *gin.Context) {
	instrumentName := context.Param("instrumentName")

	uacs, err := uacController.UacGenerator.GetAllUacsByCaseID(instrumentName)
	if err != nil {
		_ = context.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	uacs.BuildUacChunks()
	context.JSON(http.StatusOK, uacs)
}

func (uacController *UacController) ListInstrumentsEndpoint(context *gin.Context) {
	instrumentNames, err := uacController.UacGenerator.GetInstruments()
	if err != nil {
		_ = context.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	context.JSON(http.StatusOK, instrumentNames)
}

func (uacController *UacController) UACCountEndpoint(context *gin.Context) {
	instrumentName := context.Param("instrumentName")

	uacCount, err := uacController.UacGenerator.GetUacCount(instrumentName)
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
	importCount, err := uacController.UacGenerator.ImportUACs(uacs)
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
	_ = context.AbortWithError(http.StatusInternalServerError, err)
}

func (uacController *UacController) getUacRequest(context *gin.Context) (UACRequest, error) {
	body, err := io.ReadAll(context.Request.Body)
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

func (uacController *UacController) UACDisableEndpoint(context *gin.Context) {
	// uac := context.Param("uac")
	// TODO: Code below is copy/pasted, need to change
	instrumentNames, err := uacController.UacGenerator.GetInstruments()
	if err != nil {
		_ = context.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	context.JSON(http.StatusOK, instrumentNames)
}

func (uacController *UacController) UACEnableEndpoint(context *gin.Context) {
	// uac := context.Param("uac")
	// TODO: Code below is copy/pasted, need to change
	instrumentNames, err := uacController.UacGenerator.GetInstruments()
	if err != nil {
		_ = context.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	context.JSON(http.StatusOK, instrumentNames)
}

func (uacController *UacController) UACGetAllDisabledEndpoint(context *gin.Context) {
	instrumentName := context.Param("instrumentName")
	uacs, err := uacController.UacGenerator.GetAllUacsDisabled(instrumentName)
	if err != nil {
		_ = context.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	uacs.BuildUacChunks()
	context.JSON(http.StatusOK, uacs)
}
