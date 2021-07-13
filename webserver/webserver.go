package webserver

import (
	"github.com/ONSDigital/blaise-uac-service/blaiserestapi"
	"github.com/ONSDigital/blaise-uac-service/uacgenerator"
	"github.com/gin-gonic/gin"
)

type Server struct {
	BlaiseRestApi *blaiserestapi.BlaiseRestApi
	UacGenerator  *uacgenerator.UacGenerator
}

func (server *Server) SetupRouter() *gin.Engine {
	httpRouter := gin.Default()
	uacController := &UacController{
		BlaiseRestApi: server.BlaiseRestApi,
		UacGenerator:  server.UacGenerator,
	}
	uacController.AddRoutes(httpRouter)
	return httpRouter
}
