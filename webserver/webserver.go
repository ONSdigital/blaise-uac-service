package webserver

import (
	"github.com/ONSDigital/blaise-uac-service/blaiserestapi"
	"github.com/ONSDigital/blaise-uac-service/uacgenerator"
	"github.com/gin-gonic/gin"
)

type Server struct {
	BlaiseRestApi blaiserestapi.BlaiseRestApiInterface
	UacService  uacgenerator.UacServiceInterface
}

func (server *Server) SetupRouter() *gin.Engine {
	httpRouter := gin.Default()
	uacController := &UacController{
		BlaiseRestApi: server.BlaiseRestApi,
		UacService:  server.UacService,
	}
	uacController.AddRoutes(httpRouter)
	healthController := &HealthController{}
	healthController.AddRoutes(httpRouter)
	return httpRouter
}
