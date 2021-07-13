package webserver_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/ONSDigital/blaise-uac-service/blaiserestapi"
	"github.com/ONSDigital/blaise-uac-service/uacgenerator"
	"github.com/ONSDigital/blaise-uac-service/webserver"
	"github.com/gin-gonic/gin"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	mockblaiserestapi "github.com/ONSDigital/blaise-uac-service/blaiserestapi/mocks"
	mockuacgenerator "github.com/ONSDigital/blaise-uac-service/uacgenerator/mocks"
)

var _ = Describe("UAC Controller", func() {
	var (
		httpRouter    *gin.Engine
		uacController = &webserver.UacController{}
	)

	BeforeEach(func() {
		httpRouter = gin.Default()
		uacController.AddRoutes(httpRouter)
	})

	Describe("/uacs/generate/:instrumentName", func() {
		var (
			httpRecorder      *httptest.ResponseRecorder
			mockBlaiseRestApi *mockblaiserestapi.BlaiseRestApiInterface
			mockUacGenerator  *mockuacgenerator.UacGeneratorInterface
		)

		JustBeforeEach(func() {
			httpRecorder = httptest.NewRecorder()
			req, _ := http.NewRequest("POST", "/uacs/test123", nil)
			httpRouter.ServeHTTP(httpRecorder, req)
		})

		Context("when the instrument does not exist", func() {
			BeforeEach(func() {
				mockBlaiseRestApi = &mockblaiserestapi.BlaiseRestApiInterface{}

				uacController.BlaiseRestApi = mockBlaiseRestApi

				mockBlaiseRestApi.On("GetInstrumentModes", "test123").Return(blaiserestapi.InstrumentModes{}, fmt.Errorf("Instrument not found"))
			})

			It("returns a http 400 error", func() {
				Expect(httpRecorder.Code).To(Equal(http.StatusBadRequest))
				Expect(httpRecorder.Body.String()).To(Equal(`{"error":"Instrument not found"}`))
			})
		})

		Context("when the instrument does not have a CAWI mode", func() {
			BeforeEach(func() {
				mockBlaiseRestApi = &mockblaiserestapi.BlaiseRestApiInterface{}

				uacController.BlaiseRestApi = mockBlaiseRestApi

				mockBlaiseRestApi.On("GetInstrumentModes", "test123").Return(blaiserestapi.InstrumentModes{}, nil)
			})

			It("returns a http 400 error", func() {
				Expect(httpRecorder.Code).To(Equal(http.StatusBadRequest))
				Expect(httpRecorder.Body.String()).To(Equal(`{"error":"Instrument 'test123' is not installed in CAWI mode"}`))
			})
		})

		Context("when the instrument has a CAWI mode", func() {
			BeforeEach(func() {
				mockBlaiseRestApi = &mockblaiserestapi.BlaiseRestApiInterface{}
				mockUacGenerator = &mockuacgenerator.UacGeneratorInterface{}

				uacController.BlaiseRestApi = mockBlaiseRestApi
				uacController.UacGenerator = mockUacGenerator

				mockBlaiseRestApi.On("GetInstrumentModes", "test123").Return(blaiserestapi.InstrumentModes{"CAWI"}, nil)

				mockUacGenerator.On("Generate", "test123", []string{"12345"}).Return(nil)
			})

			Context("when the instrument does exist when getting case ids", func() {
				BeforeEach(func() {
					mockBlaiseRestApi.On("GetCaseIds", "test123").Return([]string{"12345"}, nil)
					mockUacGenerator.On("GetAllUacs", "test123").Return(map[string]*uacgenerator.UacInfo{
						"125634896985": &uacgenerator.UacInfo{
							InstrumentName: "test123",
							CaseID:         "12452",
						},
					}, nil)
				})

				It("generates and return a bunch of UACs", func() {
					Expect(httpRecorder.Code).To(Equal(http.StatusOK))
					Expect(httpRecorder.Body.String()).To(Equal(`{"125634896985":{"instrument_name":"test123","case_id":"12452"}}`))
				})
			})

			Context("when the instrument does not exist when getting case ids", func() {
				BeforeEach(func() {
					mockBlaiseRestApi.On("GetCaseIds", "test123").Return(nil, fmt.Errorf("Instrument not found"))
				})

				It("returns a http 400 error", func() {
					Expect(httpRecorder.Code).To(Equal(http.StatusBadRequest))
					Expect(httpRecorder.Body.String()).To(Equal(`{"error":"Instrument not found"}`))
				})
			})
		})
	})
})
