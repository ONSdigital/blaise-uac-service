package uacgenerator_test

import (
	"context"
	"math/rand"
	"time"

	"github.com/ONSDigital/blaise-uac-service/uacgenerator"
	"github.com/ONSDigital/blaise-uac-service/uacgenerator/mocks"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ = Describe("NewUac", func() {
	var (
		uacGenerator   = &uacgenerator.UacGenerator{}
		ctx            = context.Background()
		instrumentName = "lolcat"
		caseID         = "74628568"
	)

	BeforeEach(func() {
		rand.Seed(time.Now().UTC().UnixNano())
	})

	Context("Generation rules", func() {
		var mockDatastore *mocks.Datastore
		BeforeEach(func() {
			mockDatastore = &mocks.Datastore{}
			uacGenerator.DatastoreClient = mockDatastore
			uacGenerator.Context = ctx
			mockDatastore.On("Mutate",
				ctx,
				mock.AnythingOfTypeArgument("*datastore.Mutation"),
			).Return(nil, nil)
		})

		It("Generates a random 12 digit UAC", func() {
			for i := 1; i <= 20; i++ {
				uac, err := uacGenerator.NewUac(instrumentName, caseID, 0)
				Expect(uac).To(MatchRegexp(`^\d{12}$`))
				Expect(err).To(BeNil())
			}
		})
	})

	Context("when a generated UAC already exists in datastore", func() {
		var mockDatastore *mocks.Datastore
		BeforeEach(func() {
			mockDatastore = &mocks.Datastore{}
			uacGenerator.DatastoreClient = mockDatastore
			uacGenerator.Context = ctx
			mockDatastore.On("Mutate",
				ctx,
				mock.AnythingOfTypeArgument("*datastore.Mutation"),
			).Twice().Return(nil, status.Error(codes.AlreadyExists, "Already exists"))
			mockDatastore.On("Mutate",
				ctx,
				mock.AnythingOfTypeArgument("*datastore.Mutation"),
			).Return(nil, nil)
		})

		It("Regenerates a new random UAC and saves it to datastore", func() {
			uacGenerator.NewUac(instrumentName, caseID, 0)
			mockDatastore.AssertNumberOfCalls(GinkgoT(), "Mutate", 3)
		})
	})

	Context("when a generated UAC does not exist in datastore", func() {
		var mockDatastore *mocks.Datastore
		BeforeEach(func() {
			mockDatastore = &mocks.Datastore{}
			uacGenerator.DatastoreClient = mockDatastore
			uacGenerator.Context = ctx
			mockDatastore.On("Mutate",
				ctx,
				mock.AnythingOfTypeArgument("*datastore.Mutation"),
			).Return(nil, nil)
		})

		It("Saves the UAC to datastore", func() {
			uacGenerator.NewUac(instrumentName, caseID, 0)
			mockDatastore.AssertNumberOfCalls(GinkgoT(), "Mutate", 1)
		})
	})

	Context("when a generated UAC already exists in datastore over 10 times", func() {
		var mockDatastore *mocks.Datastore
		BeforeEach(func() {
			mockDatastore = &mocks.Datastore{}
			uacGenerator.DatastoreClient = mockDatastore
			uacGenerator.Context = ctx
			mockDatastore.On("Mutate",
				ctx,
				mock.AnythingOfTypeArgument("*datastore.Mutation"),
			).Return(nil, status.Error(codes.AlreadyExists, "Already exists"))
		})

		It("gives up generating a UAC and returns an error", func() {
			uac, err := uacGenerator.NewUac(instrumentName, caseID, 0)
			Expect(uac).To(Equal(""))
			mockDatastore.AssertNumberOfCalls(GinkgoT(), "Mutate", 10)
			Expect(err).To(MatchError("Could not generate a unique UAC in 10 attempts"))
		})
	})
})
