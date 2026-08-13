package webserver

import (
	"github.com/ONSDigital/blaise-uac-service/blaiserestapi"
	"github.com/ONSDigital/blaise-uac-service/uacgenerator"
	"github.com/gin-gonic/gin"
)

type Server struct {
	BlaiseRESTAPI blaiserestapi.BlaiseRESTAPIInterface
	UACService    uacgenerator.UACServiceInterface
}

func (server *Server) SetupRouter() *gin.Engine {
	httpRouter := gin.New()
	// Use panic recovery explicitly; request logging is handled by platform/infrastructure.
	httpRouter.Use(gin.Recovery())
	uacController := &UACController{
		BlaiseRESTAPI: server.BlaiseRESTAPI,
		UACService:    server.UACService,
	}
	uacController.addRoutes(httpRouter)
	healthController := &HealthController{}
	healthController.addRoutes(httpRouter)
	return httpRouter
}
