package uacgenerator_test

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"cloud.google.com/go/datastore"
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
		uacGenerator = &uacgenerator.UacGenerator{
			Context: context.Background(),
		}
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

			mockDatastore.On("Mutate",
				uacGenerator.Context,
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

			mockDatastore.On("Mutate",
				uacGenerator.Context,
				mock.AnythingOfTypeArgument("*datastore.Mutation"),
			).Twice().Return(nil, status.Error(codes.AlreadyExists, "Already exists"))
			mockDatastore.On("Mutate",
				uacGenerator.Context,
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

			mockDatastore.On("Mutate",
				uacGenerator.Context,
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

			mockDatastore.On("Mutate",
				uacGenerator.Context,
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

var _ = Describe("UacKey", func() {
	var uacGenerator = &uacgenerator.UacGenerator{}

	It("Generates a datastore named key of the correct kind", func() {
		key := uacGenerator.UacKey("test123")

		Expect(key.Kind).To(Equal(uacgenerator.UACKIND))
		Expect(key.Name).To(Equal("test123"))
	})
})

var _ = Describe("UacExistsForCase", func() {
	var (
		uacGenerator = &uacgenerator.UacGenerator{
			Context: context.Background(),
		}
		instrumentName = "lolcat"
		caseID         = "74628568"
	)

	Context("When a UAC already exists", func() {
		BeforeEach(func() {
			mockDatastore := &mocks.Datastore{}

			uacGenerator.DatastoreClient = mockDatastore

			mockDatastore.On("GetAll",
				uacGenerator.Context,
				mock.AnythingOfTypeArgument("*datastore.Query"),
				mock.AnythingOfTypeArgument("*[]*uacgenerator.UacInfo"),
			).Return([]*datastore.Key{datastore.IncompleteKey("foo", nil)}, nil)
		})

		It("returns true", func() {
			exists, err := uacGenerator.UacExistsForCase(instrumentName, caseID)

			Expect(exists).To(BeTrue())
			Expect(err).To(BeNil())
		})
	})

	Context("When a UAC does not exist", func() {
		BeforeEach(func() {
			mockDatastore := &mocks.Datastore{}

			uacGenerator.DatastoreClient = mockDatastore

			mockDatastore.On("GetAll",
				uacGenerator.Context,
				mock.AnythingOfTypeArgument("*datastore.Query"),
				mock.AnythingOfTypeArgument("*[]*uacgenerator.UacInfo"),
			).Return(nil, nil)
		})

		It("returns false", func() {
			exists, err := uacGenerator.UacExistsForCase(instrumentName, caseID)

			Expect(exists).To(BeFalse())
			Expect(err).To(BeNil())
		})
	})
})

var _ = Describe("Generate", func() {
	var (
		uacGenerator = &uacgenerator.UacGenerator{
			Context: context.Background(),
		}
		instrumentName = "lolcat"
		caseIDs        = []string{
			"74628568",
			"74628561",
			"74628562",
			"74628563",
			"74628564",
		}
		mockDatastore *mocks.Datastore
	)

	Context("when none of the cases have a uac", func() {
		BeforeEach(func() {
			mockDatastore = &mocks.Datastore{}

			uacGenerator.DatastoreClient = mockDatastore

			mockDatastore.On("GetAll",
				uacGenerator.Context,
				mock.AnythingOfTypeArgument("*datastore.Query"),
				mock.AnythingOfTypeArgument("*[]*uacgenerator.UacInfo"),
			).Return(nil, nil)

			mockDatastore.On("Mutate",
				uacGenerator.Context,
				mock.AnythingOfTypeArgument("*datastore.Mutation"),
			).Return(nil, nil)
		})

		It("generates uacs for all case ids in an instrument", func() {
			Expect(uacGenerator.Generate(instrumentName, caseIDs)).To(BeNil())

			mockDatastore.AssertNumberOfCalls(GinkgoT(), "Mutate", len(caseIDs))
			mockDatastore.AssertNumberOfCalls(GinkgoT(), "GetAll", len(caseIDs))
		})
	})

	Context("when at least one generation errors", func() {
		BeforeEach(func() {
			mockDatastore = &mocks.Datastore{}

			uacGenerator.DatastoreClient = mockDatastore

			mockDatastore.On("GetAll",
				uacGenerator.Context,
				mock.AnythingOfTypeArgument("*datastore.Query"),
				mock.AnythingOfTypeArgument("*[]*uacgenerator.UacInfo"),
			).Return(nil, nil)

			mockDatastore.On("Mutate",
				uacGenerator.Context,
				mock.AnythingOfTypeArgument("*datastore.Mutation"),
			).Once().Return(nil, nil)
			mockDatastore.On("Mutate",
				uacGenerator.Context,
				mock.AnythingOfTypeArgument("*datastore.Mutation"),
			).Once().Return(nil, fmt.Errorf("Massive mutation explosion"))
			mockDatastore.On("Mutate",
				uacGenerator.Context,
				mock.AnythingOfTypeArgument("*datastore.Mutation"),
			).Return(nil, nil)
		})

		It("returns an error", func() {
			Expect(uacGenerator.Generate(instrumentName, caseIDs)).To(MatchError("Massive mutation explosion"))
		})
	})

	Context("when one of the cases already has a uac", func() {
		BeforeEach(func() {
			mockDatastore = &mocks.Datastore{}

			uacGenerator.DatastoreClient = mockDatastore

			mockDatastore.On("GetAll",
				uacGenerator.Context,
				mock.AnythingOfTypeArgument("*datastore.Query"),
				mock.AnythingOfTypeArgument("*[]*uacgenerator.UacInfo"),
			).Once().Return([]*datastore.Key{datastore.IncompleteKey("foo", nil)}, nil)

			mockDatastore.On("GetAll",
				uacGenerator.Context,
				mock.AnythingOfTypeArgument("*datastore.Query"),
				mock.AnythingOfTypeArgument("*[]*uacgenerator.UacInfo"),
			).Return(nil, nil)

			mockDatastore.On("Mutate",
				uacGenerator.Context,
				mock.AnythingOfTypeArgument("*datastore.Mutation"),
			).Return(nil, nil)
		})

		It("generates uacs for all case ids in an instrument", func() {
			Expect(uacGenerator.Generate(instrumentName, caseIDs)).To(BeNil())

			mockDatastore.AssertNumberOfCalls(GinkgoT(), "Mutate", len(caseIDs)-1)
			mockDatastore.AssertNumberOfCalls(GinkgoT(), "GetAll", len(caseIDs))
		})
	})

	Context("when there are no cases", func() {
		BeforeEach(func() {
			mockDatastore = &mocks.Datastore{}

			uacGenerator.DatastoreClient = mockDatastore

			mockDatastore.On("GetAll",
				uacGenerator.Context,
				mock.AnythingOfTypeArgument("*datastore.Query"),
				mock.AnythingOfTypeArgument("*[]*uacgenerator.UacInfo"),
			).Return(nil, nil)

			mockDatastore.On("Mutate",
				uacGenerator.Context,
				mock.AnythingOfTypeArgument("*datastore.Mutation"),
			).Return(nil, nil)
		})

		It("generates uacs for all case ids in an instrument", func() {
			Expect(uacGenerator.Generate(instrumentName, []string{})).To(BeNil())

			mockDatastore.AssertNumberOfCalls(GinkgoT(), "Mutate", 0)
			mockDatastore.AssertNumberOfCalls(GinkgoT(), "GetAll", 0)
		})
	})
})

var _ = Describe("GetAllUacs", func() {
	var (
		uacGenerator = &uacgenerator.UacGenerator{
			Context: context.Background(),
		}
		instrumentName = "lolcat"
		mockDatastore  *mocks.Datastore
	)

	BeforeEach(func() {
		mockDatastore = &mocks.Datastore{}

		uacGenerator.DatastoreClient = mockDatastore

		mockDatastore.On("GetAll",
			uacGenerator.Context,
			mock.AnythingOfTypeArgument("*datastore.Query"),
			mock.AnythingOfTypeArgument("*[]*uacgenerator.UacInfo"),
		).Once().Return(
			func(ctx context.Context, qry *datastore.Query, dst interface{}) []*datastore.Key {
				uacInfos := dst.(*[]*uacgenerator.UacInfo)
				key := uacGenerator.UacKey("foobar")
				*uacInfos = append(*uacInfos, &uacgenerator.UacInfo{
					InstrumentName: instrumentName,
					CaseID:         "12343",
					UAC:            key,
				})
				key2 := uacGenerator.UacKey("foobar2")
				*uacInfos = append(*uacInfos, &uacgenerator.UacInfo{
					InstrumentName: instrumentName,
					CaseID:         "56764",
					UAC:            key2,
				})
				return []*datastore.Key{key, key2}
			},
			func(ctx context.Context, qry *datastore.Query, dst interface{}) error {
				return nil
			})
	})

	It("returns a map of all uacs with info", func() {
		uacs, err := uacGenerator.GetAllUacs(instrumentName)
		Expect(uacs).To(HaveLen(2))
		Expect(uacs["foobar"].InstrumentName).To(Equal(instrumentName))
		Expect(uacs["foobar"].CaseID).To(Equal("12343"))
		Expect(uacs["foobar2"].InstrumentName).To(Equal(instrumentName))
		Expect(uacs["foobar2"].CaseID).To(Equal("56764"))
		Expect(err).To(BeNil())
	})
})

var _ = Describe("GetUacCount", func() {

	var (
		uacGenerator = &uacgenerator.UacGenerator{
			Context: context.Background(),
		}
		instrumentName = "lolcat"
		mockDatastore  *mocks.Datastore
	)

	BeforeEach(func() {
		mockDatastore = &mocks.Datastore{}

		uacGenerator.DatastoreClient = mockDatastore

		mockDatastore.On("Count",
			uacGenerator.Context,
			mock.AnythingOfTypeArgument("*datastore.Query"),
		).Return(40, nil)
	})

	It("returns a map of all uacs with info", func() {
		count, err := uacGenerator.GetUacCount(instrumentName)
		Expect(count).To(Equal(40))
		Expect(err).To(BeNil())
	})
})

var _ = Describe("GetUacInfo", func() {
	var (
		uacGenerator = &uacgenerator.UacGenerator{
			Context: context.Background(),
		}
		instrumentName = "lolcat"
		mockDatastore  *mocks.Datastore
	)

	BeforeEach(func() {
		mockDatastore = &mocks.Datastore{}

		uacGenerator.DatastoreClient = mockDatastore

		mockDatastore.On("Get",
			uacGenerator.Context,
			mock.AnythingOfTypeArgument("*datastore.Key"),
			mock.AnythingOfTypeArgument("*uacgenerator.UacInfo"),
		).Once().Return(
			func(ctx context.Context, keyQry *datastore.Key, dst interface{}) error {
				uacInfo := dst.(*uacgenerator.UacInfo)
				key := uacGenerator.UacKey("lemons")
				*uacInfo = uacgenerator.UacInfo{
					InstrumentName: instrumentName,
					CaseID:         "12343",
					UAC:            key,
				}
				return nil
			})
	})

	It("Returns the uac info for a valid uac key", func() {
		uacInfo, err := uacGenerator.GetUacInfo("lemons")
		Expect(uacInfo.InstrumentName).To(Equal(instrumentName))
		Expect(uacInfo.CaseID).To(Equal("12343"))
		Expect(err).To(BeNil())
	})
})

var _ = Describe("IncrementPostcodeAttempts", func() {
	var (
		uacGenerator = &uacgenerator.UacGenerator{
			Context: context.Background(),
		}
		instrumentName = "lolcat"
		mockDatastore  *mocks.Datastore
	)

	Context("when no attempts have been made", func() {
		BeforeEach(func() {
			mockDatastore = &mocks.Datastore{}

			uacGenerator.DatastoreClient = mockDatastore

			mockDatastore.On("Get",
				uacGenerator.Context,
				mock.AnythingOfTypeArgument("*datastore.Key"),
				mock.AnythingOfTypeArgument("*uacgenerator.UacInfo"),
			).Once().Return(
				func(ctx context.Context, keyQry *datastore.Key, dst interface{}) error {
					uacInfo := dst.(*uacgenerator.UacInfo)
					key := uacGenerator.UacKey("lemons")
					*uacInfo = uacgenerator.UacInfo{
						InstrumentName: instrumentName,
						CaseID:         "12343",
						UAC:            key,
					}
					return nil
				})
			mockDatastore.On("Mutate",
				uacGenerator.Context,
				mock.AnythingOfTypeArgument("*datastore.Mutation"),
			).Return(nil, nil)
		})

		It("increments the postcode attempts and returns uacInfo", func() {
			uacInfo, err := uacGenerator.IncrementPostcodeAttempts("lemons")
			Expect(uacInfo.InstrumentName).To(Equal(instrumentName))
			Expect(uacInfo.CaseID).To(Equal("12343"))
			Expect(uacInfo.PostcodeAttempts).To(Equal(1))
			Expect(err).To(BeNil())
			layout := "2006-01-02 15:04:05.999999999 -0700 MST"
			_, err = time.Parse(layout, uacInfo.PostcodeAttemptTimestamp)
			Expect(err).To(BeNil())
		})
	})

	Context("when some attempts have been made", func() {
		BeforeEach(func() {
			mockDatastore = &mocks.Datastore{}

			uacGenerator.DatastoreClient = mockDatastore

			mockDatastore.On("Get",
				uacGenerator.Context,
				mock.AnythingOfTypeArgument("*datastore.Key"),
				mock.AnythingOfTypeArgument("*uacgenerator.UacInfo"),
			).Once().Return(
				func(ctx context.Context, keyQry *datastore.Key, dst interface{}) error {
					uacInfo := dst.(*uacgenerator.UacInfo)
					key := uacGenerator.UacKey("lemons")
					*uacInfo = uacgenerator.UacInfo{
						InstrumentName:   instrumentName,
						CaseID:           "12343",
						UAC:              key,
						PostcodeAttempts: 3,
					}
					return nil
				})
			mockDatastore.On("Mutate",
				uacGenerator.Context,
				mock.AnythingOfTypeArgument("*datastore.Mutation"),
			).Return(nil, nil)
		})

		It("increments the postcode attempts and returns uacInfo", func() {
			uacInfo, err := uacGenerator.IncrementPostcodeAttempts("lemons")
			Expect(uacInfo.InstrumentName).To(Equal(instrumentName))
			Expect(uacInfo.CaseID).To(Equal("12343"))
			Expect(uacInfo.PostcodeAttempts).To(Equal(4))
			Expect(err).To(BeNil())
			layout := "2006-01-02 15:04:05.999999999 -0700 MST"
			_, err = time.Parse(layout, uacInfo.PostcodeAttemptTimestamp)
			Expect(err).To(BeNil())
		})
	})
})

var _ = Describe("ResetPostcodeAttempts", func() {
	var (
		uacGenerator = &uacgenerator.UacGenerator{
			Context: context.Background(),
		}
		instrumentName = "lolcat"
		mockDatastore  *mocks.Datastore
	)

	Context("when no attempts have been made", func() {
		BeforeEach(func() {
			mockDatastore = &mocks.Datastore{}

			uacGenerator.DatastoreClient = mockDatastore

			mockDatastore.On("Get",
				uacGenerator.Context,
				mock.AnythingOfTypeArgument("*datastore.Key"),
				mock.AnythingOfTypeArgument("*uacgenerator.UacInfo"),
			).Once().Return(
				func(ctx context.Context, keyQry *datastore.Key, dst interface{}) error {
					uacInfo := dst.(*uacgenerator.UacInfo)
					key := uacGenerator.UacKey("lemons")
					*uacInfo = uacgenerator.UacInfo{
						InstrumentName: instrumentName,
						CaseID:         "12343",
						UAC:            key,
					}
					return nil
				})
			mockDatastore.On("Mutate",
				uacGenerator.Context,
				mock.AnythingOfTypeArgument("*datastore.Mutation"),
			).Return(nil, nil)
		})

		It("increments the postcode attempts and returns uacInfo", func() {
			uacInfo, err := uacGenerator.ResetPostcodeAttempts("lemons")
			Expect(uacInfo.InstrumentName).To(Equal(instrumentName))
			Expect(uacInfo.CaseID).To(Equal("12343"))
			Expect(uacInfo.PostcodeAttempts).To(Equal(0))
			Expect(uacInfo.PostcodeAttemptTimestamp).To(Equal(""))
			Expect(err).To(BeNil())
		})
	})

	Context("when some attempts have been made", func() {
		BeforeEach(func() {
			mockDatastore = &mocks.Datastore{}

			uacGenerator.DatastoreClient = mockDatastore

			mockDatastore.On("Get",
				uacGenerator.Context,
				mock.AnythingOfTypeArgument("*datastore.Key"),
				mock.AnythingOfTypeArgument("*uacgenerator.UacInfo"),
			).Once().Return(
				func(ctx context.Context, keyQry *datastore.Key, dst interface{}) error {
					uacInfo := dst.(*uacgenerator.UacInfo)
					key := uacGenerator.UacKey("lemons")
					*uacInfo = uacgenerator.UacInfo{
						InstrumentName:   instrumentName,
						CaseID:           "12343",
						UAC:              key,
						PostcodeAttempts: 3,
					}
					return nil
				})
			mockDatastore.On("Mutate",
				uacGenerator.Context,
				mock.AnythingOfTypeArgument("*datastore.Mutation"),
			).Return(nil, nil)
		})

		It("increments the postcode attempts and returns uacInfo", func() {
			uacInfo, err := uacGenerator.ResetPostcodeAttempts("lemons")
			Expect(uacInfo.InstrumentName).To(Equal(instrumentName))
			Expect(uacInfo.CaseID).To(Equal("12343"))
			Expect(uacInfo.PostcodeAttempts).To(Equal(0))
			Expect(uacInfo.PostcodeAttemptTimestamp).To(Equal(""))
			Expect(err).To(BeNil())
		})
	})
})
