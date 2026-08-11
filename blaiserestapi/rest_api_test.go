package blaiserestapi_test

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/ONSDigital/blaise-uac-service/blaiserestapi"
	"github.com/jarcoal/httpmock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Blaise REST API", func() {
	var (
		restAPIURL     = "http://localhost"
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
		blaiseRESTAPI = &blaiserestapi.BlaiseRESTAPI{
			BaseURL:    restAPIURL,
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

	Describe("GetCaseIDs", func() {
		Context("when an instrument does not exist", func() {
			JustBeforeEach(func() {
				httpmock.RegisterResponder("GET", fmt.Sprintf("%s/api/v2/serverparks/%s/questionnaires/%s/cases/ids", restAPIURL, serverpark, instrumentName),
					httpmock.NewBytesResponder(404, []byte{}))
			})

			It("returns a not found error", func() {
				receivedInstrumentModes, err := blaiseRESTAPI.GetCaseIDs(instrumentName)
				Expect(errors.Is(err, blaiserestapi.ErrInstrumentNotFound)).To(BeTrue())
				Expect(receivedInstrumentModes).To(BeNil())
			})
		})

		Context("when there are case IDs", func() {
			JustBeforeEach(func() {
				httpmock.RegisterResponder("GET", fmt.Sprintf("%s/api/v2/serverparks/%s/questionnaires/%s/cases/ids", restAPIURL, serverpark, instrumentName),
					httpmock.NewJsonResponderOrPanic(200, caseIDs))
			})

			It("returns the case IDs", func() {
				receivedCaseIDs, err := blaiseRESTAPI.GetCaseIDs(instrumentName)
				Expect(err).To(BeNil())
				Expect(receivedCaseIDs).To(Equal(caseIDs))
			})
		})

		Context("when there are no case IDs", func() {
			JustBeforeEach(func() {
				httpmock.RegisterResponder("GET", fmt.Sprintf("%s/api/v2/serverparks/%s/questionnaires/%s/cases/ids", restAPIURL, serverpark, instrumentName),
					httpmock.NewJsonResponderOrPanic(200, []string{}))
			})

			It("returns an empty list", func() {
				receivedCaseIDs, err := blaiseRESTAPI.GetCaseIDs(instrumentName)
				Expect(err).To(BeNil())
				Expect(receivedCaseIDs).To(BeEmpty())
			})
		})
	})

	Describe("GetInstrumentModes", func() {
		var instrumentModes = blaiserestapi.InstrumentModes{
			"CATI",
			"CAWI",
			"CAPI",
		}

		Context("when an instrument does not exist", func() {
			JustBeforeEach(func() {
				httpmock.RegisterResponder("GET", fmt.Sprintf("%s/api/v2/serverparks/%s/questionnaires/%s/modes", restAPIURL, serverpark, instrumentName),
					httpmock.NewBytesResponder(404, []byte{}))
			})

			It("returns a not found error", func() {
				receivedInstrumentModes, err := blaiseRESTAPI.GetInstrumentModes(instrumentName)
				Expect(errors.Is(err, blaiserestapi.ErrInstrumentNotFound)).To(BeTrue())
				Expect(receivedInstrumentModes).To(BeNil())
			})
		})

		Context("when an instrument has modes", func() {
			JustBeforeEach(func() {
				httpmock.RegisterResponder("GET", fmt.Sprintf("%s/api/v2/serverparks/%s/questionnaires/%s/modes", restAPIURL, serverpark, instrumentName),
					httpmock.NewJsonResponderOrPanic(200, instrumentModes))
			})

			It("returns the instrument modes", func() {
				receivedInstrumentModes, err := blaiseRESTAPI.GetInstrumentModes(instrumentName)
				Expect(err).To(BeNil())
				Expect(receivedInstrumentModes).To(Equal(instrumentModes))
			})
		})
	})
})

var _ = Describe("InstrumentModes", func() {
	Describe("HasCAWI", func() {
		Context("when the modes include CAWI", func() {
			var instrumentModes = blaiserestapi.InstrumentModes{
				"CATI",
				"CAWI",
				"CAPI",
			}

			It("returns true", func() {
				Expect(instrumentModes.HasCAWI()).To(BeTrue())
			})
		})

		Context("when the modes do not include CAWI", func() {
			var instrumentModes = blaiserestapi.InstrumentModes{
				"CATI",
				"CAPI",
			}

			It("returns false", func() {
				Expect(instrumentModes.HasCAWI()).To(BeFalse())
			})
		})
	})
})
