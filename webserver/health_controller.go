package webserver

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Health struct {
	Healthy bool `json:"healthy"`
}

type HealthController struct {
}

func (healthController *HealthController) AddRoutes(httpRouter *gin.Engine) {
	httpRouter.GET("/health", func(context *gin.Context) {
		context.JSON(http.StatusOK, Health{Healthy: true})
	})
}
