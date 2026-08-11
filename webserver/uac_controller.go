package webserver

import (
	"errors"
	"fmt"
	"net/http"

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
}

type UACController struct {
	BlaiseRESTAPI blaiserestapi.BlaiseRESTAPIInterface
	UACService    uacgenerator.UACServiceInterface
}

func (uacController *UACController) addRoutes(httpRouter *gin.Engine) {
	uacsGroup := httpRouter.Group("/uacs")
	uacsGroup.POST("/instrument/:instrumentName", uacController.generateInstrumentUACs)
	uacsGroup.GET("/instrument/:instrumentName", uacController.getAllUACs)
	uacsGroup.GET("/instrument/:instrumentName/bycaseid", uacController.getAllUACsByCaseID)
	uacsGroup.GET("/instrument/:instrumentName/count", uacController.getUACCount)
	uacsGroup.POST("/generate", uacController.generateUACs)
	// POST (not GET) so the UAC is in the request body and never appears in server access logs.
	uacsGroup.POST("/uac", uacController.getUACInfo)
	uacsGroup.DELETE("/admin/instrument/:instrumentName", uacController.deleteInstrumentUACs)
	uacsGroup.GET("/instruments", uacController.listInstruments)
	uacsGroup.POST("/import", uacController.importUACs)
	// POST (not PATCH) for the same reason: UAC stays in the body, out of access logs.
	uacsGroup.POST("/uac/disable", uacController.disableUAC)
	uacsGroup.POST("/uac/enable", uacController.enableUAC)
	uacsGroup.GET("/instrument/:instrumentName/disabled", uacController.getAllDisabledUACs)
}

func (uacController *UACController) generateInstrumentUACs(context *gin.Context) {
	instrumentName := context.Param("instrumentName")
	instrumentModes, err := uacController.BlaiseRESTAPI.GetInstrumentModes(instrumentName)
	if err != nil {
		uacController.handleError(context, err)
		return
	}
	if !instrumentModes.HasCAWI() {
		context.AbortWithStatusJSON(http.StatusBadRequest, ResponseError{Error: fmt.Sprintf("Instrument '%s' is not installed in CAWI mode", instrumentName)})
		return
	}
	caseIDs, err := uacController.BlaiseRESTAPI.GetCaseIDs(instrumentName)
	if err != nil {
		uacController.handleError(context, err)
		return
	}
	err = uacController.UACService.Generate(context.Request.Context(), instrumentName, caseIDs)
	if err != nil {
		context.AbortWithStatusJSON(http.StatusInternalServerError, ResponseError{Error: err.Error()})
		return
	}
	uacs, err := uacController.UACService.GetAllUACs(context.Request.Context(), instrumentName)
	if err != nil {
		context.AbortWithStatusJSON(http.StatusInternalServerError, ResponseError{Error: err.Error()})
		return
	}
	uacs.BuildUACChunks()
	context.JSON(http.StatusOK, uacs)
}

func (uacController *UACController) generateUACs(context *gin.Context) {
	var uacGenerateRequest UACGenerateRequest
	err := context.ShouldBindJSON(&uacGenerateRequest)
	if err != nil {
		context.AbortWithStatusJSON(http.StatusBadRequest, ResponseError{Error: err.Error()})
		return
	}
	if uacGenerateRequest.InstrumentName == "" {
		context.AbortWithStatusJSON(http.StatusBadRequest, ResponseError{Error: "Must provide instrument name"})
		return
	}
	err = uacController.UACService.Generate(context.Request.Context(), uacGenerateRequest.InstrumentName, uacGenerateRequest.CaseIDs)
	if err != nil {
		context.AbortWithStatusJSON(http.StatusInternalServerError, ResponseError{Error: err.Error()})
		return
	}
	uacs, err := uacController.UACService.GetAllUACs(context.Request.Context(), uacGenerateRequest.InstrumentName)
	if err != nil {
		context.AbortWithStatusJSON(http.StatusInternalServerError, ResponseError{Error: err.Error()})
		return
	}
	uacs.BuildUACChunks()
	context.JSON(http.StatusOK, uacs)
}

func (uacController *UACController) getAllUACs(context *gin.Context) {
	instrumentName := context.Param("instrumentName")

	uacs, err := uacController.UACService.GetAllUACs(context.Request.Context(), instrumentName)
	if err != nil {
		context.AbortWithStatusJSON(http.StatusInternalServerError, ResponseError{Error: err.Error()})
		return
	}
	uacs.BuildUACChunks()
	context.JSON(http.StatusOK, uacs)
}

func (uacController *UACController) getAllUACsByCaseID(context *gin.Context) {
	instrumentName := context.Param("instrumentName")

	uacs, err := uacController.UACService.GetAllUACsByCaseID(context.Request.Context(), instrumentName)
	if err != nil {
		context.AbortWithStatusJSON(http.StatusInternalServerError, ResponseError{Error: err.Error()})
		return
	}
	uacs.BuildUACChunks()
	context.JSON(http.StatusOK, uacs)
}

func (uacController *UACController) listInstruments(context *gin.Context) {
	instrumentNames, err := uacController.UACService.GetInstruments(context.Request.Context())
	if err != nil {
		context.AbortWithStatusJSON(http.StatusInternalServerError, ResponseError{Error: err.Error()})
		return
	}
	context.JSON(http.StatusOK, instrumentNames)
}

func (uacController *UACController) getUACCount(context *gin.Context) {
	instrumentName := context.Param("instrumentName")

	uacCount, err := uacController.UACService.GetUACCount(context.Request.Context(), instrumentName)
	if err != nil {
		context.AbortWithStatusJSON(http.StatusInternalServerError, ResponseError{Error: err.Error()})
		return
	}
	context.JSON(http.StatusOK, gin.H{"count": uacCount})
}

func (uacController *UACController) getUACInfo(context *gin.Context) {
	var req UACRequest
	if err := context.ShouldBindJSON(&req); err != nil {
		context.AbortWithStatusJSON(http.StatusBadRequest, ResponseError{Error: err.Error()})
		return
	}
	uac := req.UAC

	uacInfo, err := uacController.UACService.GetUACInfo(context.Request.Context(), uac)
	if err != nil {
		if errors.Is(err, uacgenerator.ErrNotFound) {
			context.AbortWithStatusJSON(http.StatusNotFound, ResponseError{Error: err.Error()})
			return
		}
		context.AbortWithStatusJSON(http.StatusInternalServerError, ResponseError{Error: err.Error()})
		return
	}
	context.JSON(http.StatusOK, uacInfo)
}

func (uacController *UACController) deleteInstrumentUACs(context *gin.Context) {
	instrumentName := context.Param("instrumentName")
	err := uacController.UACService.AdminDelete(context.Request.Context(), instrumentName)
	if err != nil {
		context.AbortWithStatusJSON(http.StatusInternalServerError, ResponseError{Error: err.Error()})
		return
	}
	context.Status(http.StatusNoContent)
}

func (uacController *UACController) importUACs(context *gin.Context) {
	var uacs []string
	err := context.ShouldBindJSON(&uacs)
	if err != nil {
		context.AbortWithStatusJSON(http.StatusBadRequest, ResponseError{Error: err.Error()})
		return
	}
	importCount, err := uacController.UACService.ImportUACs(context.Request.Context(), uacs)
	if err != nil {
		var importErr *uacgenerator.ImportError
		if errors.As(err, &importErr) {
			context.AbortWithStatusJSON(http.StatusBadRequest, ResponseError{Error: err.Error()})
			return
		}
		_ = context.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	context.JSON(http.StatusOK, gin.H{"uacs_imported": importCount})
}

func (uacController *UACController) handleError(context *gin.Context, err error) {
	if errors.Is(err, blaiserestapi.ErrInstrumentNotFound) {
		context.AbortWithStatusJSON(http.StatusBadRequest, ResponseError{Error: err.Error()})
		return
	}
	context.AbortWithStatusJSON(http.StatusInternalServerError, ResponseError{Error: err.Error()})
}

func (uacController *UACController) disableUAC(context *gin.Context) {
	var req UACRequest
	if err := context.ShouldBindJSON(&req); err != nil {
		context.AbortWithStatusJSON(http.StatusBadRequest, ResponseError{Error: err.Error()})
		return
	}
	uac := req.UAC

	err := uacController.UACService.DisableUAC(context.Request.Context(), uac)
	if err != nil {
		if errors.Is(err, uacgenerator.ErrInvalidUAC) {
			context.AbortWithStatusJSON(http.StatusBadRequest, ResponseError{Error: err.Error()})
			return
		}
		context.AbortWithStatusJSON(http.StatusInternalServerError, ResponseError{Error: err.Error()})
		return
	}
	context.Status(http.StatusNoContent)
}

func (uacController *UACController) enableUAC(context *gin.Context) {
	var req UACRequest
	if err := context.ShouldBindJSON(&req); err != nil {
		context.AbortWithStatusJSON(http.StatusBadRequest, ResponseError{Error: err.Error()})
		return
	}
	uac := req.UAC

	err := uacController.UACService.EnableUAC(context.Request.Context(), uac)
	if err != nil {
		if errors.Is(err, uacgenerator.ErrInvalidUAC) {
			context.AbortWithStatusJSON(http.StatusBadRequest, ResponseError{Error: err.Error()})
			return
		}
		context.AbortWithStatusJSON(http.StatusInternalServerError, ResponseError{Error: err.Error()})
		return
	}
	context.Status(http.StatusNoContent)
}

func (uacController *UACController) getAllDisabledUACs(context *gin.Context) {
	instrumentName := context.Param("instrumentName")

	uacs, err := uacController.UACService.GetAllDisabledUACs(context.Request.Context(), instrumentName)
	if err != nil {
		uacController.handleError(context, err)
		return
	}
	uacs.BuildUACChunks()
	context.JSON(http.StatusOK, uacs)
}
