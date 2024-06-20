package webserver_test

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"

	"cloud.google.com/go/datastore"
	"github.com/ONSDigital/blaise-uac-service/blaiserestapi"
	"github.com/ONSDigital/blaise-uac-service/uacgenerator"
	"github.com/ONSDigital/blaise-uac-service/webserver"
	"github.com/gin-gonic/gin"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/mock"

	mockblaiserestapi "github.com/ONSDigital/blaise-uac-service/blaiserestapi/mocks"
	mockuacgenerator "github.com/ONSDigital/blaise-uac-service/uacgenerator/mocks"
)

var _ = Describe("UAC Controller", func() {
	var (
		httpRouter        *gin.Engine
		mockBlaiseRestApi = &mockblaiserestapi.BlaiseRestApiInterface{}
		mockUacGenerator  = &mockuacgenerator.UacGeneratorInterface{}
		uacController     = &webserver.UacController{UacGenerator: mockUacGenerator, BlaiseRestApi: mockBlaiseRestApi}
	)

	BeforeEach(func() {
		httpRouter = gin.Default()
		uacController.AddRoutes(httpRouter)
	})

	AfterEach(func() {
		mockBlaiseRestApi = &mockblaiserestapi.BlaiseRestApiInterface{}
		mockUacGenerator = &mockuacgenerator.UacGeneratorInterface{}
		uacController.UacGenerator = mockUacGenerator
		uacController.BlaiseRestApi = mockBlaiseRestApi
	})

	Describe("POST /uacs/instrument/:instrumentName", func() {
		var (
			httpRecorder *httptest.ResponseRecorder
		)

		JustBeforeEach(func() {
			httpRecorder = httptest.NewRecorder()
			req, _ := http.NewRequest("POST", "/uacs/instrument/test123", nil)
			httpRouter.ServeHTTP(httpRecorder, req)
		})

		Context("when the instrument does not exist", func() {
			BeforeEach(func() {
				mockBlaiseRestApi.On("GetInstrumentModes", "test123").Return(blaiserestapi.InstrumentModes{}, fmt.Errorf("Instrument not found"))
			})

			It("returns a http 400 error", func() {
				Expect(httpRecorder.Code).To(Equal(http.StatusBadRequest))
				Expect(httpRecorder.Body.String()).To(Equal(`{"error":"Instrument not found"}`))
			})
		})

		Context("when the instrument does not have a CAWI mode", func() {
			BeforeEach(func() {
				mockBlaiseRestApi.On("GetInstrumentModes", "test123").Return(blaiserestapi.InstrumentModes{}, nil)
			})

			It("returns a http 400 error", func() {
				Expect(httpRecorder.Code).To(Equal(http.StatusBadRequest))
				Expect(httpRecorder.Body.String()).To(Equal(`{"error":"Instrument 'test123' is not installed in CAWI mode"}`))
			})
		})

		Context("when the instrument has a CAWI mode", func() {
			BeforeEach(func() {
				mockBlaiseRestApi.On("GetInstrumentModes", "test123").Return(blaiserestapi.InstrumentModes{"CAWI"}, nil)

				mockUacGenerator.On("Generate", "test123", []string{"12345"}).Return(nil)
			})

			Context("when the instrument does exist when getting case ids", func() {
				BeforeEach(func() {
					mockBlaiseRestApi.On("GetCaseIds", "test123").Return([]string{"12345"}, nil)
					mockUacGenerator.On("GetAllUacs", "test123").Return(uacgenerator.Uacs{
						"125634896985": {
							InstrumentName: "test123",
							CaseID:         "12452",
						},
					}, nil)
				})

				It("generates and return a bunch of UACs", func() {
					Expect(httpRecorder.Code).To(Equal(http.StatusOK))
					Expect(httpRecorder.Body.String()).To(Equal(`{"125634896985":{"instrument_name":"test123","case_id":"12452","uac_chunks":{"uac1":"1256","uac2":"3489","uac3":"6985"},"disabled":false}}`))
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

	Describe("GET /uacs/instrument/:instrumentName", func() {
		var (
			httpRecorder *httptest.ResponseRecorder
		)

		JustBeforeEach(func() {
			httpRecorder = httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/uacs/instrument/test123", nil)
			httpRouter.ServeHTTP(httpRecorder, req)
		})

		Context("When the instrument has UAC codes", func() {
			BeforeEach(func() {
				mockUacGenerator.On("GetAllUacs", "test123").Return(uacgenerator.Uacs{
					"125634896985": {
						InstrumentName: "test123",
						CaseID:         "12452",
					},
					"78945612309": {
						InstrumentName: "test123",
						CaseID:         "65858",
					},
				}, nil)
			})

			It("Gets all UACs for an installed instrument", func() {
				Expect(httpRecorder.Code).To(Equal(http.StatusOK))
				Expect(httpRecorder.Body.String()).To(Equal(`{"125634896985":{"instrument_name":"test123","case_id":"12452","uac_chunks":{"uac1":"1256","uac2":"3489","uac3":"6985"},"disabled":false},"78945612309":{"instrument_name":"test123","case_id":"65858","uac_chunks":{"uac1":"7894","uac2":"5612","uac3":"309"},"disabled":false}}`))
			})
		})

		Context("When the instrument has UAC Info held against it", func() {
			BeforeEach(func() {
				mockUacGenerator.On("GetAllUacs", "test123").Return(uacgenerator.Uacs{}, nil)
			})

			It("Returns an empty list with status code of Ok", func() {
				Expect(httpRecorder.Code).To(Equal(http.StatusOK))
				Expect(httpRecorder.Body.String()).To(Equal(`{}`))
			})
		})
	})

	Describe("GET /uacs/instrument/:instrumentName/bycaseid", func() {
		var (
			httpRecorder *httptest.ResponseRecorder
		)

		JustBeforeEach(func() {
			httpRecorder = httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/uacs/instrument/test123/bycaseid", nil)
			httpRouter.ServeHTTP(httpRecorder, req)
		})

		BeforeEach(func() {
			mockUacGenerator.On("GetAllUacsByCaseID", "test123").Return(uacgenerator.Uacs{
				"12452": {
					InstrumentName: "test123",
					CaseID:         "12452",
					FullUAC:        "125634896985",
				},
				"65858": {
					InstrumentName: "test123",
					CaseID:         "65858",
					FullUAC:        "78945612309",
				},
			}, nil)
		})

		It("Gets all UACs for an installed instrument", func() {
			Expect(httpRecorder.Code).To(Equal(http.StatusOK))
			Expect(httpRecorder.Body.String()).To(Equal(`{"12452":{"instrument_name":"test123","case_id":"12452","uac_chunks":{"uac1":"1256","uac2":"3489","uac3":"6985"},"full_uac":"125634896985","disabled":false},"65858":{"instrument_name":"test123","case_id":"65858","uac_chunks":{"uac1":"7894","uac2":"5612","uac3":"309"},"full_uac":"78945612309","disabled":false}}`))
		})
	})

	Describe("/uacs/instrument/:instrumentName/count", func() {
		var (
			httpRecorder *httptest.ResponseRecorder
		)

		JustBeforeEach(func() {
			httpRecorder = httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/uacs/instrument/test123/count", nil)
			httpRouter.ServeHTTP(httpRecorder, req)
		})

		BeforeEach(func() {
			mockUacGenerator.On("GetUacCount", "test123").Return(20, nil)
		})

		It("Returns a number of uacs with a status Ok", func() {
			Expect(httpRecorder.Code).To(Equal(http.StatusOK))
			Expect(httpRecorder.Body.String()).To(Equal(`{"count":20}`))
		})
	})

	Describe("/uacs/instruments", func() {
		var (
			httpRecorder *httptest.ResponseRecorder
		)

		JustBeforeEach(func() {
			httpRecorder = httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/uacs/instruments", nil)
			httpRouter.ServeHTTP(httpRecorder, req)
		})

		BeforeEach(func() {
			mockUacGenerator.On("GetInstruments").Return([]string{"foo", "bar"}, nil)
		})

		It("Returns instrument_names with a status Ok", func() {
			Expect(httpRecorder.Code).To(Equal(http.StatusOK))
			Expect(httpRecorder.Body.String()).To(Equal(`["foo","bar"]`))
		})
	})

	Describe("POST /uacs/generate", func() {
		var (
			httpRecorder *httptest.ResponseRecorder
		)

		Context("when case_ids and instrument_name are provided", func() {
			JustBeforeEach(func() {
				requestBody := `{"instrument_name": "test123", "case_ids": ["123", "456", "789"]}`
				httpRecorder = httptest.NewRecorder()
				req, _ := http.NewRequest("POST", "/uacs/generate", bytes.NewBufferString(requestBody))
				httpRouter.ServeHTTP(httpRecorder, req)
			})

			BeforeEach(func() {
				mockUacGenerator.On("Generate", "test123", []string{"123", "456", "789"}).Return(nil)
				mockUacGenerator.On("GetAllUacs", "test123").Return(uacgenerator.Uacs{
					"125634896985": {
						InstrumentName: "test123",
						CaseID:         "12452",
					},
				}, nil)
			})

			It("generates and return a bunch of UACs", func() {
				Expect(httpRecorder.Code).To(Equal(http.StatusOK))
				Expect(httpRecorder.Body.String()).To(Equal(`{"125634896985":{"instrument_name":"test123","case_id":"12452","uac_chunks":{"uac1":"1256","uac2":"3489","uac3":"6985"},"disabled":false}}`))
			})
		})

		Context("when no case_ids are provided", func() {
			JustBeforeEach(func() {
				requestBody := `{"instrument_name": "test123"}`
				httpRecorder = httptest.NewRecorder()
				req, _ := http.NewRequest("POST", "/uacs/generate", bytes.NewBufferString(requestBody))
				httpRouter.ServeHTTP(httpRecorder, req)
			})

			BeforeEach(func() {
				mockUacGenerator.On("Generate", "test123", []string(nil)).Return(nil)
				mockUacGenerator.On("GetAllUacs", "test123").Return(uacgenerator.Uacs{}, nil)
			})

			It("generated nothing, and returns as such", func() {
				Expect(httpRecorder.Code).To(Equal(http.StatusOK))
				Expect(httpRecorder.Body.String()).To(Equal(`{}`))
			})
		})

		Context("when instrument_name is not provided", func() {
			JustBeforeEach(func() {
				requestBody := `{"case_ids": ["123", "456", "789"]}`
				httpRecorder = httptest.NewRecorder()
				req, _ := http.NewRequest("POST", "/uacs/generate", bytes.NewBufferString(requestBody))
				httpRouter.ServeHTTP(httpRecorder, req)
			})

			It("generates and return a bunch of UACs", func() {
				Expect(httpRecorder.Code).To(Equal(http.StatusBadRequest))
				Expect(httpRecorder.Body.String()).To(Equal(`{"error":"Must provide instrument name"}`))
			})
		})
	})

	Describe("/uacs/uac", func() {
		var (
			httpRecorder *httptest.ResponseRecorder
			requestBody  io.Reader
		)

		JustBeforeEach(func() {
			httpRecorder = httptest.NewRecorder()
			req, _ := http.NewRequest("POST", "/uacs/uac", requestBody)
			httpRouter.ServeHTTP(httpRecorder, req)
		})

		Context("A valid UAC returns UACInfo for that code", func() {
			BeforeEach(func() {
				requestBody = bytes.NewReader([]byte(`{"uac":"98765432101"}`))
				mockUacGenerator.On("GetUacInfo", "98765432101").Return(&uacgenerator.UacInfo{
					InstrumentName: "test123",
					CaseID:         "12452",
				}, nil)
			})

			It("Gets UAC Info for a valid UAC Code", func() {
				Expect(httpRecorder.Code).To(Equal(http.StatusOK))
				Expect(httpRecorder.Body.String()).To(Equal(`{"instrument_name":"test123","case_id":"12452","disabled":false}`))
			})
		})

		Context("Returns bad request if no body is posted", func() {
			BeforeEach(func() {
				requestBody = bytes.NewReader([]byte(``))
			})

			It("Returns an empty body and a bad request status", func() {
				Expect(httpRecorder.Code).To(Equal(http.StatusBadRequest))
				Expect(httpRecorder.Body.String()).To(Equal("null"))
			})
		})

		Context("Returns bad request if no body is invalid JSON", func() {
			BeforeEach(func() {
				requestBody = bytes.NewReader([]byte(`{"blah":Blah}`))
			})

			It("Returns an empty body and a bad request status", func() {
				Expect(httpRecorder.Code).To(Equal(http.StatusBadRequest))
				Expect(httpRecorder.Body.String()).To(Equal("null"))
			})
		})

		Context("Returns bad request if no body is invalid JSON", func() {
			BeforeEach(func() {
				requestBody = bytes.NewReader([]byte(`{"uac":"98765432101"}`))
				mockUacGenerator.On("GetUacInfo", "98765432101").Return(nil, datastore.ErrNoSuchEntity)
			})

			It("Returns an empty body and a not found status", func() {
				Expect(httpRecorder.Code).To(Equal(http.StatusNotFound))
				Expect(httpRecorder.Body.String()).To(Equal("null"))
			})
		})
	})

	Describe("POST /import", func() {
		var (
			httpRecorder *httptest.ResponseRecorder
		)

		JustBeforeEach(func() {
			requestBody := `["123456789123","123456789145","123556789987"]`
			httpRecorder = httptest.NewRecorder()
			req, _ := http.NewRequest("POST", "/uacs/import", bytes.NewBufferString(requestBody))
			httpRouter.ServeHTTP(httpRecorder, req)
		})

		Context("and importing the UACs is successful", func() {
			BeforeEach(func() {
				mockUacGenerator.On("ImportUACs", mock.AnythingOfType("[]string")).Return(3, nil)
			})

			It("imports all of the UACs", func() {
				Expect(httpRecorder.Code).To(Equal(http.StatusOK))
				Expect(httpRecorder.Body.String()).To(Equal(`{"uacs_imported":3}`))
			})
		})

		Context("and importing the UACs errors", func() {
			Context("and the error is an import error", func() {
				BeforeEach(func() {
					mockUacGenerator.On("ImportUACs", mock.AnythingOfType("[]string")).
						Return(0, &uacgenerator.ImportError{InvalidUACs: []string{"foobar"}})
				})

				It("errors and doesn't import anything", func() {
					Expect(httpRecorder.Code).To(Equal(http.StatusBadRequest))
					Expect(httpRecorder.Body.String()).To(Equal(`{"error":"Cannot import UACs because some were invalid: [\"foobar\"]"}`))
				})
			})

			Context("and the error is any other error", func() {
				BeforeEach(func() {
					mockUacGenerator.On("ImportUACs", mock.AnythingOfType("[]string")).Return(0, fmt.Errorf("invalid uac"))
				})

				It("errors and doesn't import anything", func() {
					Expect(httpRecorder.Code).To(Equal(http.StatusInternalServerError))
				})
			})
		})
	})

	Describe("GET /uacs/uac/:instrumentName/disabled", func() {
		var (
			httpRecorder *httptest.ResponseRecorder
		)

		JustBeforeEach(func() {
			httpRecorder = httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/uacs/uac/test123/disabled", nil)
			httpRouter.ServeHTTP(httpRecorder, req)
		})

		BeforeEach(func() {
			mockUacGenerator.On("GetAllUacsDisabled", "test123").Return(uacgenerator.Uacs{
				"12452": {
					InstrumentName: "test123",
					CaseID:         "12452",
					FullUAC:        "125634896985",
					Disabled:       true,
				},
				"65858": {
					InstrumentName: "test123",
					CaseID:         "65858",
					FullUAC:        "78945612309",
					Disabled:       true,
				},
			}, nil)
		})

		It("Gets all UACs for an installed instrument that are disabled", func() {
			Expect(httpRecorder.Code).To(Equal(http.StatusOK))
			// TODO: put back in Expect(httpRecorder.Body.String()).To(Equal(`{"12452":{"instrument_name":"test123","case_id":"12452","uac_chunks":{"uac1":"1256","uac2":"3489","uac3":"6985"},"full_uac":"125634896985","disabled":true},"65858":{"instrument_name":"test123","case_id":"65858","uac_chunks":{"uac1":"7894","uac2":"5612","uac3":"309"},"full_uac":"78945612309","disabled":true}}`))
		})
	})
})
