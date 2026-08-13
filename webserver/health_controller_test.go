package webserver_test

import (
	"net/http"
	"net/http/httptest"

	"github.com/ONSDigital/blaise-uac-service/webserver"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Server router and health endpoints", func() {
	It("serves health and versioned health routes", func() {
		server := &webserver.Server{}
		httpRouter := server.SetupRouter()

		healthRecorder := httptest.NewRecorder()
		healthReq, _ := http.NewRequest("GET", "/health", nil)
		httpRouter.ServeHTTP(healthRecorder, healthReq)

		Expect(healthRecorder.Code).To(Equal(http.StatusOK))
		Expect(healthRecorder.Body.String()).To(Equal(`{"healthy":true}`))

		versionedRecorder := httptest.NewRecorder()
		versionedReq, _ := http.NewRequest("GET", "/bus/v2/health", nil)
		httpRouter.ServeHTTP(versionedRecorder, versionedReq)

		Expect(versionedRecorder.Code).To(Equal(http.StatusOK))
		Expect(versionedRecorder.Body.String()).To(Equal(`{"healthy":true,"version":"v2"}`))
	})
})
