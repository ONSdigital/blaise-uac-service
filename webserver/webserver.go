package webserver

import "github.com/gin-gonic/gin"

func SetupRouter() *gin.Engine {
	httpRouter := gin.Default()
	UACRoutes(httpRouter)
	return httpRouter
}
