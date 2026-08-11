package webserver_test

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/ONSDigital/blaise-uac-service/blaiserestapi"
	"github.com/ONSDigital/blaise-uac-service/uacgenerator"
	"github.com/ONSDigital/blaise-uac-service/webserver"
	"github.com/gin-gonic/gin"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/mock"

	mockblaiserestapi "github.com/ONSDigital/blaise-uac-service/blaiserestapi/mocks"
	mockuacgenerator "github.com/ONSDigital/blaise-uac-service/uacgenerator/mocks"
)

var _ = Describe("UAC controller", func() {
	var (
		httpRouter        *gin.Engine
		mockBlaiseRESTAPI = &mockblaiserestapi.BlaiseRESTAPIInterface{}
		mockUACService    = &mockuacgenerator.UACServiceInterface{}
	)

	BeforeEach(func() {
		server := &webserver.Server{UACService: mockUACService, BlaiseRESTAPI: mockBlaiseRESTAPI}
		httpRouter = server.SetupRouter()
	})

	AfterEach(func() {
		mockBlaiseRESTAPI = &mockblaiserestapi.BlaiseRESTAPIInterface{}
		mockUACService = &mockuacgenerator.UACServiceInterface{}
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
				mockBlaiseRESTAPI.On("GetInstrumentModes", "test123").Return(blaiserestapi.InstrumentModes{}, blaiserestapi.ErrInstrumentNotFound)
			})

			It("returns a http 400 error", func() {
				Expect(httpRecorder.Code).To(Equal(http.StatusBadRequest))
				Expect(httpRecorder.Body.String()).To(Equal(`{"error":"instrument not found"}`))
			})
		})

		Context("when the instrument does not have a CAWI mode", func() {
			BeforeEach(func() {
				mockBlaiseRESTAPI.On("GetInstrumentModes", "test123").Return(blaiserestapi.InstrumentModes{}, nil)
			})

			It("returns a http 400 error", func() {
				Expect(httpRecorder.Code).To(Equal(http.StatusBadRequest))
				Expect(httpRecorder.Body.String()).To(Equal(`{"error":"Instrument 'test123' is not installed in CAWI mode"}`))
			})
		})

		Context("when the instrument has a CAWI mode", func() {
			BeforeEach(func() {
				mockBlaiseRESTAPI.On("GetInstrumentModes", "test123").Return(blaiserestapi.InstrumentModes{"CAWI"}, nil)

				mockUACService.On("Generate", mock.Anything, "test123", []string{"12345"}).Return(nil)
			})

			Context("when the instrument does exist when getting case ids", func() {
				BeforeEach(func() {
					mockBlaiseRESTAPI.On("GetCaseIDs", "test123").Return([]string{"12345"}, nil)
					mockUACService.On("GetAllUACs", mock.Anything, "test123").Return(uacgenerator.UACs{
						"125634896985": {
							InstrumentName: "test123",
							CaseID:         "12452",
						},
					}, nil)
				})

				It("generates and returns UACs for the instrument", func() {
					Expect(httpRecorder.Code).To(Equal(http.StatusOK))
					Expect(httpRecorder.Body.String()).To(Equal(`{"125634896985":{"instrument_name":"test123","case_id":"12452","uac_chunks":{"uac1":"1256","uac2":"3489","uac3":"6985"},"disabled":false}}`))
				})
			})

			Context("when the instrument does not exist when getting case ids", func() {
				BeforeEach(func() {
					mockBlaiseRESTAPI.On("GetCaseIDs", "test123").Return(nil, blaiserestapi.ErrInstrumentNotFound)
				})

				It("returns a http 400 error", func() {
					Expect(httpRecorder.Code).To(Equal(http.StatusBadRequest))
					Expect(httpRecorder.Body.String()).To(Equal(`{"error":"instrument not found"}`))
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

		Context("when the instrument has UACs", func() {
			BeforeEach(func() {
				mockUACService.On("GetAllUACs", mock.Anything, "test123").Return(uacgenerator.UACs{
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

			It("returns all UACs for the instrument", func() {
				Expect(httpRecorder.Code).To(Equal(http.StatusOK))
				Expect(httpRecorder.Body.String()).To(Equal(`{"125634896985":{"instrument_name":"test123","case_id":"12452","uac_chunks":{"uac1":"1256","uac2":"3489","uac3":"6985"},"disabled":false},"78945612309":{"instrument_name":"test123","case_id":"65858","uac_chunks":{"uac1":"7894","uac2":"5612","uac3":"309"},"disabled":false}}`))
			})
		})

		Context("when the instrument has no UACs", func() {
			BeforeEach(func() {
				mockUACService.On("GetAllUACs", mock.Anything, "test123").Return(uacgenerator.UACs{}, nil)
			})

			It("returns an empty response", func() {
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
			mockUACService.On("GetAllUACsByCaseID", mock.Anything, "test123").Return(uacgenerator.UACs{
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

		It("returns all UACs keyed by case ID", func() {
			Expect(httpRecorder.Code).To(Equal(http.StatusOK))
			Expect(httpRecorder.Body.String()).To(Equal(`{"12452":{"instrument_name":"test123","case_id":"12452","uac_chunks":{"uac1":"1256","uac2":"3489","uac3":"6985"},"full_uac":"125634896985","disabled":false},"65858":{"instrument_name":"test123","case_id":"65858","uac_chunks":{"uac1":"7894","uac2":"5612","uac3":"309"},"full_uac":"78945612309","disabled":false}}`))
		})
	})

	Describe("GET /uacs/instrument/:instrumentName/count", func() {
		var (
			httpRecorder *httptest.ResponseRecorder
		)

		JustBeforeEach(func() {
			httpRecorder = httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/uacs/instrument/test123/count", nil)
			httpRouter.ServeHTTP(httpRecorder, req)
		})

		BeforeEach(func() {
			mockUACService.On("GetUACCount", mock.Anything, "test123").Return(20, nil)
		})

		It("returns the UAC count", func() {
			Expect(httpRecorder.Code).To(Equal(http.StatusOK))
			Expect(httpRecorder.Body.String()).To(Equal(`{"count":20}`))
		})
	})

	Describe("GET /uacs/instruments", func() {
		var (
			httpRecorder *httptest.ResponseRecorder
		)

		JustBeforeEach(func() {
			httpRecorder = httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/uacs/instruments", nil)
			httpRouter.ServeHTTP(httpRecorder, req)
		})

		BeforeEach(func() {
			mockUACService.On("GetInstruments", mock.Anything).Return([]string{"foo", "bar"}, nil)
		})

		It("returns the instrument names", func() {
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
				mockUACService.On("Generate", mock.Anything, "test123", []string{"123", "456", "789"}).Return(nil)
				mockUACService.On("GetAllUACs", mock.Anything, "test123").Return(uacgenerator.UACs{
					"125634896985": {
						InstrumentName: "test123",
						CaseID:         "12452",
					},
				}, nil)
			})

			It("generates and returns the UACs", func() {
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
				mockUACService.On("Generate", mock.Anything, "test123", []string(nil)).Return(nil)
				mockUACService.On("GetAllUACs", mock.Anything, "test123").Return(uacgenerator.UACs{}, nil)
			})

			It("returns an empty response", func() {
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

			It("returns a bad request error", func() {
				Expect(httpRecorder.Code).To(Equal(http.StatusBadRequest))
				Expect(httpRecorder.Body.String()).To(Equal(`{"error":"Must provide instrument name"}`))
			})
		})

		Context("when the request body is malformed JSON", func() {
			JustBeforeEach(func() {
				requestBody := `{"instrument_name": "test123",`
				httpRecorder = httptest.NewRecorder()
				req, _ := http.NewRequest("POST", "/uacs/generate", bytes.NewBufferString(requestBody))
				httpRouter.ServeHTTP(httpRecorder, req)
			})

			It("returns a bad request before generation is attempted", func() {
				Expect(httpRecorder.Code).To(Equal(http.StatusBadRequest))
				mockUACService.AssertNotCalled(GinkgoT(), "Generate", mock.Anything, mock.Anything, mock.Anything)
			})
		})
	})

	Describe("GET /uacs/uac/:uac", func() {
		var (
			httpRecorder *httptest.ResponseRecorder
		)

		JustBeforeEach(func() {
			httpRecorder = httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/uacs/uac/98765432101", nil)
			httpRouter.ServeHTTP(httpRecorder, req)
		})

		Context("when the UAC is valid", func() {
			BeforeEach(func() {
				mockUACService.On("GetUACInfo", mock.Anything, "98765432101").Return(&uacgenerator.UACInfo{
					InstrumentName: "test123",
					CaseID:         "12452",
				}, nil)
			})

			It("returns the UAC info", func() {
				Expect(httpRecorder.Code).To(Equal(http.StatusOK))
				Expect(httpRecorder.Body.String()).To(Equal(`{"instrument_name":"test123","case_id":"12452","disabled":false}`))
			})
		})

		Context("when the UAC does not exist", func() {
			BeforeEach(func() {
				mockUACService.On("GetUACInfo", mock.Anything, "98765432101").Return(nil, uacgenerator.ErrNotFound)
			})

			It("returns a not found status", func() {
				Expect(httpRecorder.Code).To(Equal(http.StatusNotFound))
				Expect(httpRecorder.Body.String()).To(Equal(`{"error":"not found"}`))
			})
		})
	})

	Describe("POST /uacs/import", func() {
		var (
			httpRecorder *httptest.ResponseRecorder
			requestBody  string
		)

		JustBeforeEach(func() {
			httpRecorder = httptest.NewRecorder()
			req, _ := http.NewRequest("POST", "/uacs/import", bytes.NewBufferString(requestBody))
			httpRouter.ServeHTTP(httpRecorder, req)
		})

		BeforeEach(func() {
			requestBody = `["123456789123","123456789145","123556789987"]`
		})

		Context("and importing the UACs is successful", func() {
			BeforeEach(func() {
				mockUACService.On("ImportUACs", mock.Anything, mock.AnythingOfType("[]string")).Return(3, nil)
			})

			It("imports all of the UACs", func() {
				Expect(httpRecorder.Code).To(Equal(http.StatusOK))
				Expect(httpRecorder.Body.String()).To(Equal(`{"uacs_imported":3}`))
			})
		})

		Context("and importing the UACs errors", func() {
			Context("and the error is an import error", func() {
				BeforeEach(func() {
					mockUACService.On("ImportUACs", mock.Anything, mock.AnythingOfType("[]string")).
						Return(0, &uacgenerator.ImportError{InvalidUACs: []string{"foobar"}})
				})

				It("errors and doesn't import anything", func() {
					Expect(httpRecorder.Code).To(Equal(http.StatusBadRequest))
					Expect(httpRecorder.Body.String()).To(Equal(`{"error":"Cannot import UACs because some were invalid: [\"foobar\"]"}`))
				})
			})

			Context("and the error wraps an import error", func() {
				BeforeEach(func() {
					wrappedErr := fmt.Errorf("wrapped import error: %w", &uacgenerator.ImportError{InvalidUACs: []string{"foobar"}})
					mockUACService.On("ImportUACs", mock.Anything, mock.AnythingOfType("[]string")).Return(0, wrappedErr)
				})

				It("returns a bad request", func() {
					Expect(httpRecorder.Code).To(Equal(http.StatusBadRequest))
					Expect(httpRecorder.Body.String()).To(Equal(`{"error":"wrapped import error: Cannot import UACs because some were invalid: [\"foobar\"]"}`))
				})
			})

			Context("and the error is any other error", func() {
				BeforeEach(func() {
					mockUACService.On("ImportUACs", mock.Anything, mock.AnythingOfType("[]string")).Return(0, fmt.Errorf("invalid uac"))
				})

				It("errors and doesn't import anything", func() {
					Expect(httpRecorder.Code).To(Equal(http.StatusInternalServerError))
				})
			})
		})

		Context("and the request body is malformed JSON", func() {
			BeforeEach(func() {
				requestBody = `["123456789123",`
			})

			It("returns a bad request before import is attempted", func() {
				Expect(httpRecorder.Code).To(Equal(http.StatusBadRequest))
				mockUACService.AssertNotCalled(GinkgoT(), "ImportUACs", mock.Anything)
			})
		})
	})

	Describe("DELETE /uacs/admin/instrument/:instrumentName", func() {
		var (
			httpRecorder *httptest.ResponseRecorder
		)

		JustBeforeEach(func() {
			httpRecorder = httptest.NewRecorder()
			req, _ := http.NewRequest("DELETE", "/uacs/admin/instrument/test123", nil)
			httpRouter.ServeHTTP(httpRecorder, req)
		})

		Context("when deletion succeeds", func() {
			BeforeEach(func() {
				mockUACService.On("AdminDelete", mock.Anything, "test123").Return(nil)
			})

			It("returns no content", func() {
				Expect(httpRecorder.Code).To(Equal(http.StatusNoContent))
				Expect(httpRecorder.Body.String()).To(Equal(""))
			})
		})

		Context("when deletion fails", func() {
			BeforeEach(func() {
				mockUACService.On("AdminDelete", mock.Anything, "test123").Return(fmt.Errorf("delete failed"))
			})

			It("returns an internal server error", func() {
				Expect(httpRecorder.Code).To(Equal(http.StatusInternalServerError))
				Expect(httpRecorder.Body.String()).To(Equal(`{"error":"delete failed"}`))
			})
		})
	})

	Describe("GET /uacs/instrument/:instrumentName/disabled", func() {
		var (
			httpRecorder *httptest.ResponseRecorder
		)

		JustBeforeEach(func() {
			httpRecorder = httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/uacs/instrument/test123/disabled", nil)
			httpRouter.ServeHTTP(httpRecorder, req)
		})

		BeforeEach(func() {
			mockUACService.On("GetAllDisabledUACs", mock.Anything, "test123").Return(uacgenerator.UACs{
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

		It("returns all disabled UACs for the instrument", func() {
			Expect(httpRecorder.Code).To(Equal(http.StatusOK))
			Expect(httpRecorder.Body.String()).To(Equal(`{"12452":{"instrument_name":"test123","case_id":"12452","uac_chunks":{"uac1":"1256","uac2":"3489","uac3":"6985"},"full_uac":"125634896985","disabled":true},"65858":{"instrument_name":"test123","case_id":"65858","uac_chunks":{"uac1":"7894","uac2":"5612","uac3":"309"},"full_uac":"78945612309","disabled":true}}`))
		})
	})

	Describe("PATCH /uacs/disable/:uac", func() {
		var (
			httpRecorder *httptest.ResponseRecorder
		)

		JustBeforeEach(func() {
			httpRecorder = httptest.NewRecorder()
			req, _ := http.NewRequest("PATCH", "/uacs/uac/disable/123456789", nil)
			httpRouter.ServeHTTP(httpRecorder, req)
		})

		BeforeEach(func() {
			mockUACService.On("DisableUAC", mock.Anything, "123456789").Return(nil)
		})

		It("sets the disabled flag to true", func() {
			Expect(httpRecorder.Code).To(Equal(http.StatusNoContent))
		})
	})

	Describe("PATCH /uacs/enable/:uac", func() {
		var (
			httpRecorder *httptest.ResponseRecorder
		)

		JustBeforeEach(func() {
			httpRecorder = httptest.NewRecorder()
			req, _ := http.NewRequest("PATCH", "/uacs/uac/enable/87654321", nil)
			httpRouter.ServeHTTP(httpRecorder, req)
		})

		BeforeEach(func() {
			mockUACService.On("EnableUAC", mock.Anything, "87654321").Return(nil)
		})

		It("sets the disabled flag to false", func() {
			Expect(httpRecorder.Code).To(Equal(http.StatusNoContent))
		})
	})

	Describe("PATCH /uacs/enable/:uac with a non existing uac", func() {
		var (
			httpRecorder *httptest.ResponseRecorder
		)

		JustBeforeEach(func() {
			httpRecorder = httptest.NewRecorder()
			req, _ := http.NewRequest("PATCH", "/uacs/uac/enable/1234", nil)
			httpRouter.ServeHTTP(httpRecorder, req)
		})

		BeforeEach(func() {
			mockUACService.On("EnableUAC", mock.Anything, "1234").Return(uacgenerator.ErrInvalidUAC)
		})

		It("returns a http 400 error", func() {
			Expect(httpRecorder.Code).To(Equal(http.StatusBadRequest))
			Expect(httpRecorder.Body.String()).To(Equal(`{"error":"invalid uac"}`))
		})
	})

	Describe("PATCH /uacs/enable/:uac when enabling fails unexpectedly", func() {
		var (
			httpRecorder *httptest.ResponseRecorder
		)

		JustBeforeEach(func() {
			httpRecorder = httptest.NewRecorder()
			req, _ := http.NewRequest("PATCH", "/uacs/uac/enable/1234", nil)
			httpRouter.ServeHTTP(httpRecorder, req)
		})

		BeforeEach(func() {
			mockUACService.On("EnableUAC", mock.Anything, "1234").Return(fmt.Errorf("boom"))
		})

		It("returns a http 500 error", func() {
			Expect(httpRecorder.Code).To(Equal(http.StatusInternalServerError))
			Expect(httpRecorder.Body.String()).To(Equal(`{"error":"boom"}`))
		})
	})

	Describe("PATCH /uacs/disable/:uac with a non existing uac", func() {
		var (
			httpRecorder *httptest.ResponseRecorder
		)

		JustBeforeEach(func() {
			httpRecorder = httptest.NewRecorder()
			req, _ := http.NewRequest("PATCH", "/uacs/uac/disable/1234", nil)
			httpRouter.ServeHTTP(httpRecorder, req)
		})

		BeforeEach(func() {
			mockUACService.On("DisableUAC", mock.Anything, "1234").Return(uacgenerator.ErrInvalidUAC)
		})

		It("returns a http 400 error", func() {
			Expect(httpRecorder.Code).To(Equal(http.StatusBadRequest))
			Expect(httpRecorder.Body.String()).To(Equal(`{"error":"invalid uac"}`))
		})
	})

	Describe("PATCH /uacs/disable/:uac when disabling fails unexpectedly", func() {
		var (
			httpRecorder *httptest.ResponseRecorder
		)

		JustBeforeEach(func() {
			httpRecorder = httptest.NewRecorder()
			req, _ := http.NewRequest("PATCH", "/uacs/uac/disable/1234", nil)
			httpRouter.ServeHTTP(httpRecorder, req)
		})

		BeforeEach(func() {
			mockUACService.On("DisableUAC", mock.Anything, "1234").Return(fmt.Errorf("boom"))
		})

		It("returns a http 500 error", func() {
			Expect(httpRecorder.Code).To(Equal(http.StatusInternalServerError))
			Expect(httpRecorder.Body.String()).To(Equal(`{"error":"boom"}`))
		})
	})

	Describe("GET /uacs/instrument/:instrumentName/disabled with a non existing instrumentName", func() {
		var (
			httpRecorder *httptest.ResponseRecorder
		)

		JustBeforeEach(func() {
			httpRecorder = httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/uacs/instrument/unknownInstrumentName/disabled", nil)
			httpRouter.ServeHTTP(httpRecorder, req)
		})

		BeforeEach(func() {
			mockUACService.On("GetAllDisabledUACs", mock.Anything, "unknownInstrumentName").Return(nil, blaiserestapi.ErrInstrumentNotFound)
		})

		It("returns a http 400 error", func() {
			Expect(httpRecorder.Code).To(Equal(http.StatusBadRequest))
			Expect(httpRecorder.Body.String()).To(Equal(`{"error":"instrument not found"}`))
		})
	})
})
