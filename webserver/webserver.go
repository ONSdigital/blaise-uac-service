package webserver

import (
	"github.com/ONSDigital/blaise-uac-service/blaiserestapi"
	"github.com/ONSDigital/blaise-uac-service/uacgenerator"
	"github.com/gin-gonic/gin"
)

type Server struct {
	BlaiseRestApi    blaiserestapi.BlaiseRestApiInterface
	UacGenerator     uacgenerator.UacGeneratorInterface
	TracerMiddleWare gin.HandlerFunc
}

func (server *Server) SetupRouter() *gin.Engine {
	httpRouter := gin.Default()
	httpRouter.Use(server.TracerMiddleWare)
	uacController := &UacController{
		BlaiseRestApi: server.BlaiseRestApi,
		UacGenerator:  server.UacGenerator,
	}
	uacController.AddRoutes(httpRouter)
	healthController := &HealthController{}
	healthController.AddRoutes(httpRouter)
	return httpRouter
}
