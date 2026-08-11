package webserver

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Health struct {
	Healthy bool   `json:"healthy"`
	Version string `json:"version,omitempty"`
}

type HealthController struct {
}

func (healthController *HealthController) addRoutes(httpRouter *gin.Engine) {
	httpRouter.GET("/health", healthController.getHealth)
	httpRouter.GET("/bus/:version/health", healthController.getHealth)
}

func (healthController *HealthController) getHealth(context *gin.Context) {
	version := context.Param("version")
	context.JSON(http.StatusOK, Health{Healthy: true, Version: version})
}
