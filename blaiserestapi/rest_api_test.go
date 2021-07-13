package blaiserestapi_test

import (
	"fmt"
	"net/http"

	"github.com/ONSDigital/blaise-uac-service/blaiserestapi"
	"github.com/jarcoal/httpmock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Blaise rest api endpoints", func() {
	var (
		restApiUrl     = "http://localhost"
		serverpark     = "foobar"
		instrumentName = "lolcats"
		caseIDs        = []string{
			"12345",
			"12346",
			"12347",
			"12341",
			"12344",
			"12342",
		}
		blaiseRestApi = &blaiserestapi.BlaiseRestApi{
			BaseUrl:    restApiUrl,
			Serverpark: serverpark,
			Client:     &http.Client{},
		}
	)

	BeforeEach(func() {
		httpmock.Activate()
	})

	AfterEach(func() {
		httpmock.DeactivateAndReset()
	})

	Describe("Get Case Ids", func() {
		Context("when an instrument does not exist", func() {
			JustBeforeEach(func() {
				httpmock.RegisterResponder("GET", fmt.Sprintf("%s/api/v1/serverparks/%s/instruments/%s/cases/ids", restApiUrl, serverpark, instrumentName),
					httpmock.NewBytesResponder(404, []byte{}))
			})

			It("returns a NotFound error", func() {
				recievedInstrumentModes, err := blaiseRestApi.GetCaseIds(instrumentName)
				Expect(err).To(MatchError("Instrument not found"))
				Expect(recievedInstrumentModes).To(BeNil())
			})
		})

		Context("when there are case IDs", func() {
			JustBeforeEach(func() {
				httpmock.RegisterResponder("GET", fmt.Sprintf("%s/api/v1/serverparks/%s/instruments/%s/cases/ids", restApiUrl, serverpark, instrumentName),
					httpmock.NewJsonResponderOrPanic(200, caseIDs))
			})

			It("When I call the Blaise Rest Api Case Id end point, a list of Case Ids are returned", func() {
				receivedCaseIds, err := blaiseRestApi.GetCaseIds(instrumentName)
				Expect(err).To(BeNil())
				Expect(receivedCaseIds).To(Equal(caseIDs))
			})
		})

		Context("when there are no case IDs", func() {
			JustBeforeEach(func() {
				httpmock.RegisterResponder("GET", fmt.Sprintf("%s/api/v1/serverparks/%s/instruments/%s/cases/ids", restApiUrl, serverpark, instrumentName),
					httpmock.NewJsonResponderOrPanic(200, []string{}))
			})

			It("When I call the Blaise Rest Api Case Id end point, a list of Case Ids are returned", func() {
				receivedCaseIds, err := blaiseRestApi.GetCaseIds(instrumentName)
				Expect(err).To(BeNil())
				Expect(receivedCaseIds).To(BeEmpty())
			})
		})
	})

	Describe("Get a list of modes", func() {
		var instrumentModes = blaiserestapi.InstrumentModes{
			"CATI",
			"CAWI",
			"CAPI",
		}

		Context("when an instrument does not exist", func() {
			JustBeforeEach(func() {
				httpmock.DefaultTransport.RegisterResponder("GET", fmt.Sprintf("%s/api/v1/serverparks/%s/instruments/%s/modes", restApiUrl, serverpark, instrumentName),
					httpmock.NewBytesResponder(404, []byte{}))
			})

			It("returns a NotFound error", func() {
				recievedInstrumentModes, err := blaiseRestApi.GetInstrumentModes(instrumentName)
				Expect(err).To(MatchError("Instrument not found"))
				Expect(recievedInstrumentModes).To(BeNil())
			})
		})

		Context("when an instrument has modes", func() {
			JustBeforeEach(func() {
				httpmock.DefaultTransport.RegisterResponder("GET", fmt.Sprintf("%s/api/v1/serverparks/%s/instruments/%s/modes", restApiUrl, serverpark, instrumentName),
					httpmock.NewJsonResponderOrPanic(200, instrumentModes))
			})

			It("When I call the Blaise Rest Api Modes end point, a list of modes are returned", func() {
				recievedInstrumentModes, err := blaiseRestApi.GetInstrumentModes(instrumentName)
				Expect(err).To(BeNil())
				Expect(recievedInstrumentModes).To(Equal(instrumentModes))
			})
		})
	})
})

var _ = Describe("InstrumentModes", func() {
	Describe("HasCawi", func() {
		Context("when the modes include CAWI", func() {
			var instrumentModes = blaiserestapi.InstrumentModes{
				"CATI",
				"CAWI",
				"CAPI",
			}

			It("returns true", func() {
				Expect(instrumentModes.HasCawi()).To(BeTrue())
			})
		})

		Context("when the modes do not include CAWI", func() {
			var instrumentModes = blaiserestapi.InstrumentModes{
				"CATI",
				"CAPI",
			}

			It("returns false", func() {
				Expect(instrumentModes.HasCawi()).To(BeFalse())
			})
		})
	})
})
