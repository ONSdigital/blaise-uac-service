package webserver_test

import (
	"net/http"
	"net/http/httptest"

	"github.com/ONSDigital/blaise-uac-service/webserver"
	"github.com/gin-gonic/gin"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("UAC Controller", func() {
	var (
		httpRouter   *gin.Engine
		uacControler *webserver.UacController
	)

	BeforeEach(func() {
		httpRouter = gin.Default()
		uacControler.AddRoutes(httpRouter)
	})

	Describe("/uacs/generate/:instrumentName", func() {
		var (
			httpRecorder *httptest.ResponseRecorder
		)

		BeforeEach(func() {
			httpRecorder = httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/uacs/generate/test123", nil)
			httpRouter.ServeHTTP(httpRecorder, req)
		})

		It("Generates a bunch of UACs for a given instrument", func() {
			Expect(httpRecorder.Code).To(Equal(http.StatusOK))
			Expect(httpRecorder.Body.String()).To(Equal("null"))
		})
	})
})
