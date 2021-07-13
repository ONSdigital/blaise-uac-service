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
})
