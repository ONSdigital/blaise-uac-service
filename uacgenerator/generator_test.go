package uacgenerator_test

import (
	"context"
	"fmt"
	"strconv"

	"cloud.google.com/go/datastore"
	"github.com/ONSDigital/blaise-uac-service/uacgenerator"
	"github.com/ONSDigital/blaise-uac-service/uacgenerator/mocks"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ = Describe("GenerateUac12", func() {
	var (
		uacGenerator  *uacgenerator.UacGenerator
		mockDatastore *mocks.Datastore
	)

	BeforeEach(func() {
		uacGenerator = uacgenerator.NewUacGenerator(mockDatastore, "uac")
	})

	It("Generates a random 12 digit UAC", func() {
		for i := 1; i <= 20; i++ {
			uac := uacGenerator.GenerateUac12()

			Expect(uac).To(MatchRegexp(`^\d{12}$`))

			var startIndex = 0
			for i := 0; i < 3; i++ {
				uacSegmant, _ := strconv.Atoi(uac[startIndex : startIndex+4])
				Expect(uacSegmant).To(BeNumerically(">=", 1000))
				Expect(uacSegmant).To(BeNumerically("<=", 9999))
				startIndex = startIndex + 4
			}
		}
	})
})

var _ = Describe("GenerateUac16", func() {
	var (
		uacGenerator  *uacgenerator.UacGenerator
		mockDatastore *mocks.Datastore
	)

	BeforeEach(func() {
		uacGenerator = uacgenerator.NewUacGenerator(mockDatastore, "uac16")
	})

	It("Generates a random 16 alphanumeric UAC", func() {
		var unapprovedCharacters = "aeiouyw01"
		for i := 1; i <= 20; i++ {
			uac := uacGenerator.GenerateUac16()

			Expect(uac).To(MatchRegexp(fmt.Sprintf(`^[%s]{16}$`, uacgenerator.APPROVEDCHARACTERS)))

			//Ensure does not contain unapproved characters
			Expect(uac).ToNot(MatchRegexp(fmt.Sprintf(`^.*[%s]{1}.*$`, unapprovedCharacters)))
		}
	})
})

var _ = Describe("NewUac", func() {
	var (
		uacGenerator   *uacgenerator.UacGenerator
		instrumentName = "lolcat"
		caseID         = "74628568"
	)

	Context("Generation rules for 12 digit UAC", func() {
		var mockDatastore *mocks.Datastore

		BeforeEach(func() {
			mockDatastore = &mocks.Datastore{}

			uacGenerator = uacgenerator.NewUacGenerator(mockDatastore, "uac")

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

	Context("Generation rules for 16 alphanumeric UAC", func() {
		var mockDatastore *mocks.Datastore

		BeforeEach(func() {
			mockDatastore = &mocks.Datastore{}

			uacGenerator = uacgenerator.NewUacGenerator(mockDatastore, "uac16")

			mockDatastore.On("Mutate",
				uacGenerator.Context,
				mock.AnythingOfTypeArgument("*datastore.Mutation"),
			).Return(nil, nil)
		})

		It("Generates a random 16 character alphanumeric UAC", func() {
			for i := 1; i <= 20; i++ {
				uac, err := uacGenerator.NewUac(instrumentName, caseID, 0)

				Expect(uac).To(MatchRegexp(fmt.Sprintf(`^[%s]{16}$`, uacgenerator.APPROVEDCHARACTERS)))
				Expect(err).To(BeNil())
			}
		})
	})

	Context("when a UAC kind is blank", func() {
		var mockDatastore *mocks.Datastore

		BeforeEach(func() {
			mockDatastore = &mocks.Datastore{}

			uacGenerator = uacgenerator.NewUacGenerator(mockDatastore, "")

			mockDatastore.On("Mutate",
				uacGenerator.Context,
				mock.AnythingOfTypeArgument("*datastore.Mutation"),
			).Return(nil, nil)
		})

		It("returns an error", func() {
			uac, err := uacGenerator.NewUac(instrumentName, caseID, 0)
			Expect(uac).To(BeEmpty())
			Expect(err).To(MatchError("Cannot generate UACs for invalid UacKind"))
		})
	})

	Context("when a UAC kind is invalid", func() {
		var mockDatastore *mocks.Datastore

		BeforeEach(func() {
			mockDatastore = &mocks.Datastore{}

			uacGenerator = uacgenerator.NewUacGenerator(mockDatastore, "this is not a valid UWACKY")

			mockDatastore.On("Mutate",
				uacGenerator.Context,
				mock.AnythingOfTypeArgument("*datastore.Mutation"),
			).Return(nil, nil)
		})

		It("returns an error", func() {
			uac, err := uacGenerator.NewUac(instrumentName, caseID, 0)
			Expect(uac).To(BeEmpty())
			Expect(err).To(MatchError("Cannot generate UACs for invalid UacKind"))
		})
	})

	Context("when a caseID is blank", func() {
		It("returns an error", func() {
			uacGenerator.UacKind = "uac"
			uac, err := uacGenerator.NewUac(instrumentName, "", 0)
			Expect(uac).To(BeEmpty())
			Expect(err).To(MatchError("Cannot generate UACs for blank caseIDs"))
		})
	})

	Context("when a generated UAC already exists in datastore", func() {
		var mockDatastore *mocks.Datastore
		BeforeEach(func() {
			mockDatastore = &mocks.Datastore{}

			uacGenerator = uacgenerator.NewUacGenerator(mockDatastore, "uac")

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
			_, err := uacGenerator.NewUac(instrumentName, caseID, 0)
            Expect(err).ShouldNot(HaveOccurred())
			mockDatastore.AssertNumberOfCalls(GinkgoT(), "Mutate", 3)
		})
	})

	Context("when a generated UAC does not exist in datastore", func() {
		var mockDatastore *mocks.Datastore
		BeforeEach(func() {
			mockDatastore = &mocks.Datastore{}

			uacGenerator = uacgenerator.NewUacGenerator(mockDatastore, "uac")

			mockDatastore.On("Mutate",
				uacGenerator.Context,
				mock.AnythingOfTypeArgument("*datastore.Mutation"),
			).Return(nil, nil)
		})

		It("Saves the UAC to datastore", func() {
			_, err := uacGenerator.NewUac(instrumentName, caseID, 0)
            Expect(err).ShouldNot(HaveOccurred())
			mockDatastore.AssertNumberOfCalls(GinkgoT(), "Mutate", 1)
		})
	})

	Context("when a generated UAC already exists in datastore over 10 times", func() {
		var mockDatastore *mocks.Datastore
		BeforeEach(func() {
			mockDatastore = &mocks.Datastore{}

			uacGenerator = uacgenerator.NewUacGenerator(mockDatastore, "uac")

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

		Expect(key.Kind).To(Equal(uacGenerator.UacKind))
		Expect(key.Name).To(Equal("test123"))
	})
})

var _ = Describe("UacExistsForCase", func() {
	var (
		uacGenerator   *uacgenerator.UacGenerator
		instrumentName = "lolcat"
		caseID         = "74628568"
	)

	Context("When a UAC already exists", func() {
		BeforeEach(func() {
			mockDatastore := &mocks.Datastore{}

			uacGenerator = uacgenerator.NewUacGenerator(mockDatastore, "uac")

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

			uacGenerator = uacgenerator.NewUacGenerator(mockDatastore, "uac")

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
		uacGenerator   *uacgenerator.UacGenerator
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

			uacGenerator = uacgenerator.NewUacGenerator(mockDatastore, "uac")

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

			uacGenerator = uacgenerator.NewUacGenerator(mockDatastore, "uac")

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

			uacGenerator = uacgenerator.NewUacGenerator(mockDatastore, "uac")

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

			uacGenerator = uacgenerator.NewUacGenerator(mockDatastore, "uac")

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
		uacGenerator   *uacgenerator.UacGenerator
		instrumentName = "lolcat"
		mockDatastore  *mocks.Datastore
	)

	BeforeEach(func() {
		mockDatastore = &mocks.Datastore{}

		uacGenerator = uacgenerator.NewUacGenerator(mockDatastore, "uac")

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

var _ = Describe("GetAllUacs", func() {
	var (
		uacGenerator   *uacgenerator.UacGenerator
		instrumentName = "lolcat"
		mockDatastore  *mocks.Datastore
	)

	Context("when there are duplicate case ids", func() {
		BeforeEach(func() {
			mockDatastore = &mocks.Datastore{}

			uacGenerator = uacgenerator.NewUacGenerator(mockDatastore, "uac")

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
						CaseID:         "12343",
						UAC:            key2,
					})
					return []*datastore.Key{key, key2}
				},
				func(ctx context.Context, qry *datastore.Query, dst interface{}) error {
					return nil
				})
		})

		It("returns an error", func() {
			uacs, err := uacGenerator.GetAllUacsByCaseID(instrumentName)
			Expect(uacs).To(BeNil())
			Expect(err).To(MatchError("Fewer case ids than uacs, must be duplicate case ids"))
		})
	})

	Context("when there are no duplicate case ids", func() {
		BeforeEach(func() {
			mockDatastore = &mocks.Datastore{}

			uacGenerator = uacgenerator.NewUacGenerator(mockDatastore, "uac")

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
			uacs, err := uacGenerator.GetAllUacsByCaseID(instrumentName)
			Expect(uacs).To(HaveLen(2))
			Expect(uacs["12343"].InstrumentName).To(Equal(instrumentName))
			Expect(uacs["12343"].CaseID).To(Equal("12343"))
			Expect(uacs["56764"].InstrumentName).To(Equal(instrumentName))
			Expect(uacs["56764"].CaseID).To(Equal("56764"))
			Expect(err).To(BeNil())
		})
	})
})

var _ = Describe("GetUacCount", func() {
	var (
		uacGenerator   *uacgenerator.UacGenerator
		instrumentName = "lolcat"
		mockDatastore  *mocks.Datastore
	)

	BeforeEach(func() {
		mockDatastore = &mocks.Datastore{}

		uacGenerator = uacgenerator.NewUacGenerator(mockDatastore, "uac")

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
		uacGenerator   *uacgenerator.UacGenerator
		instrumentName = "lolcat"
		mockDatastore  *mocks.Datastore
	)

	BeforeEach(func() {
		mockDatastore = &mocks.Datastore{}

		uacGenerator = uacgenerator.NewUacGenerator(mockDatastore, "uac")

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

var _ = Describe("GetInstruments", func() {
	var (
		uacGenerator  *uacgenerator.UacGenerator
		mockDatastore *mocks.Datastore
	)

	BeforeEach(func() {
		mockDatastore = &mocks.Datastore{}

		uacGenerator = uacgenerator.NewUacGenerator(mockDatastore, "uac")

		mockDatastore.On("GetAll",
			uacGenerator.Context,
			mock.AnythingOfTypeArgument("*datastore.Query"),
			mock.AnythingOfTypeArgument("*[]*uacgenerator.UacInfo"),
		).Once().Return(
			func(ctx context.Context, qry *datastore.Query, dst interface{}) []*datastore.Key {
				uacInfos := dst.(*[]*uacgenerator.UacInfo)
				*uacInfos = append(*uacInfos, &uacgenerator.UacInfo{
					InstrumentName: "foo",
				})
				*uacInfos = append(*uacInfos, &uacgenerator.UacInfo{
					InstrumentName: "bar",
				})
				return []*datastore.Key{}
			},
			func(ctx context.Context, qry *datastore.Query, dst interface{}) error {
				return nil
			})
	})

	It("Returns a list of instrument names", func() {
		instrumentNames, err := uacGenerator.GetInstruments()
		Expect(err).To(BeNil())
		Expect(instrumentNames).To(Equal([]string{"foo", "bar"}))
	})
})

var _ = DescribeTable("ChunkUAC",
	func(uac string, expected uacgenerator.UacChunks) {
		Expect(*uacgenerator.ChunkUAC(uac)).To(Equal(expected))
	},
	Entry("123456781234", "123456781234", uacgenerator.UacChunks{UAC1: "1234", UAC2: "5678", UAC3: "1234"}),
	Entry("111122223333", "111122223333", uacgenerator.UacChunks{UAC1: "1111", UAC2: "2222", UAC3: "3333"}),
	Entry("11112222333344444", "1111222233334444", uacgenerator.UacChunks{UAC1: "1111", UAC2: "2222", UAC3: "3333", UAC4: "4444"}),
)

var _ = Describe("Uacs", func() {
	Describe("BuildUacChunks", func() {
		var uacs = uacgenerator.Uacs{
			"111122223333": &uacgenerator.UacInfo{},
			"123456781234": &uacgenerator.UacInfo{},
		}

		It("Adds UacChunks to the UacInfo", func() {
			uacs.BuildUacChunks()
			Expect(*uacs["111122223333"].UacChunks).To(Equal(uacgenerator.UacChunks{UAC1: "1111", UAC2: "2222", UAC3: "3333"}))
			Expect(*uacs["123456781234"].UacChunks).To(Equal(uacgenerator.UacChunks{UAC1: "1234", UAC2: "5678", UAC3: "1234"}))
		})
	})
})

var _ = Describe("ImportUacs", func() {
	var (
		uacGenerator  *uacgenerator.UacGenerator
		mockDatastore *mocks.Datastore
		uacs          []string
	)

	BeforeEach(func() {
		mockDatastore = &mocks.Datastore{}
		uacGenerator = uacgenerator.NewUacGenerator(mockDatastore, "uac")

		mockDatastore.On("Mutate",
			uacGenerator.Context,
			mock.AnythingOfTypeArgument("*datastore.Mutation"),
		).Return(nil, nil)
	})

	AfterEach(func() {
		uacs = []string{}
	})

	Context("when there are no uacs", func() {
		BeforeEach(func() {
			uacs = []string{}
		})

		It("imports nothing and returns 0 imported with no error", func() {
			updateCount, err := uacGenerator.ImportUACs(uacs)
			Expect(updateCount).To(Equal(0))
			Expect(err).To(BeNil())
			mockDatastore.AssertNumberOfCalls(GinkgoT(), "Mutate", 0)
		})
	})

	Context("and none of the UACs already exist", func() {
		Context("and all the UACs are valid", func() {
			BeforeEach(func() {
				uacs = []string{"123456789123", "123456789145", "123556789987"}

				mockDatastore.On("Get",
					uacGenerator.Context,
					mock.AnythingOfTypeArgument("*datastore.Key"),
					mock.AnythingOfTypeArgument("*uacgenerator.UacInfo"),
				).Return(datastore.ErrNoSuchEntity)
			})

			It("imports all of the UACs", func() {
				updateCount, err := uacGenerator.ImportUACs(uacs)
				Expect(updateCount).To(Equal(3))
				Expect(err).To(BeNil())
				mockDatastore.AssertNumberOfCalls(GinkgoT(), "Mutate", 3)
			})
		})

		Context("and one of the UACs has an invalid format", func() {
			BeforeEach(func() {
				uacs = []string{"123456789123", "123456789145", "123556789987", "a2sad", "2131asda91298"}
			})

			It("errors and doesn't import anything", func() {
				updateCount, err := uacGenerator.ImportUACs(uacs)
				Expect(updateCount).To(Equal(0))
				Expect(err).To(MatchError(`Cannot import UACs because some were invalid: ["a2sad", "2131asda91298"]`))
				mockDatastore.AssertNumberOfCalls(GinkgoT(), "Mutate", 0)
			})
		})
	})

	Context("and all of the UACs already exist as 'unknown'", func() {
		BeforeEach(func() {
			uacs = []string{"123456789123", "123456789145", "123556789987"}

			mockDatastore.On("Get",
				uacGenerator.Context,
				mock.AnythingOfTypeArgument("*datastore.Key"),
				mock.AnythingOfTypeArgument("*uacgenerator.UacInfo"),
			).Return(func(ctx context.Context, keyQry *datastore.Key, dst interface{}) error {
				uacInfo := dst.(*uacgenerator.UacInfo)
				key := uacGenerator.UacKey("any")
				*uacInfo = uacgenerator.UacInfo{
					InstrumentName: "unknown",
					CaseID:         "unknown",
					UAC:            key,
				}
				return nil
			})
		})

		It("imports nothing and returns 0 imported with no error", func() {
			updateCount, err := uacGenerator.ImportUACs(uacs)
			Expect(updateCount).To(Equal(0))
			Expect(err).To(BeNil())
			mockDatastore.AssertNumberOfCalls(GinkgoT(), "Mutate", 0)
		})
	})

	Context("and some of the UACs already exist", func() {
		BeforeEach(func() {
			uacs = []string{"123456789123", "123456789145", "123556789987"}
		})

		Context("and they have an InstrumentName of 'unknown'", func() {
			BeforeEach(func() {
				mockDatastore.On("Get",
					uacGenerator.Context,
					mock.AnythingOfTypeArgument("*datastore.Key"),
					mock.AnythingOfTypeArgument("*uacgenerator.UacInfo"),
				).Times(2).Return(datastore.ErrNoSuchEntity)
				mockDatastore.On("Get",
					uacGenerator.Context,
					mock.AnythingOfTypeArgument("*datastore.Key"),
					mock.AnythingOfTypeArgument("*uacgenerator.UacInfo"),
				).Return(func(ctx context.Context, keyQry *datastore.Key, dst interface{}) error {
					uacInfo := dst.(*uacgenerator.UacInfo)
					key := uacGenerator.UacKey("123556789987")
					*uacInfo = uacgenerator.UacInfo{
						InstrumentName: "unknown",
						CaseID:         "unknown",
						UAC:            key,
					}
					return nil
				})
			})

			It("imports all of the UACs, skipping those that already exist", func() {
				updateCount, err := uacGenerator.ImportUACs(uacs)
				Expect(updateCount).To(Equal(2))
				Expect(err).To(BeNil())
				mockDatastore.AssertNumberOfCalls(GinkgoT(), "Mutate", 2)
			})
		})

		Context("and they have InstrumentNames that are not 'unknown'", func() {
			BeforeEach(func() {
				mockDatastore.On("Get",
					uacGenerator.Context,
					uacGenerator.UacKey("123556789987"),
					mock.AnythingOfTypeArgument("*uacgenerator.UacInfo"),
				).Return(func(ctx context.Context, keyQry *datastore.Key, dst interface{}) error {
					uacInfo := dst.(*uacgenerator.UacInfo)
					key := uacGenerator.UacKey("123556789987")
					*uacInfo = uacgenerator.UacInfo{
						InstrumentName: "dst2108a",
						CaseID:         "1234",
						UAC:            key,
					}
					return nil
				})
				mockDatastore.On("Get",
					uacGenerator.Context,
					mock.AnythingOfTypeArgument("*datastore.Key"),
					mock.AnythingOfTypeArgument("*uacgenerator.UacInfo"),
				).Return(datastore.ErrNoSuchEntity)
			})

			It("errors and doesn't import anything", func() {
				updateCount, err := uacGenerator.ImportUACs(uacs)
				Expect(updateCount).To(Equal(0))
				Expect(err).To(MatchError(`Cannot import UACs because some were already in use by questionnaires: ["123556789987"]`))
				mockDatastore.AssertNumberOfCalls(GinkgoT(), "Mutate", 0)
			})
		})
	})
})

var _ = Describe("ValidateUAC12", func() {
	var uacGenerator = &uacgenerator.UacGenerator{}
	DescribeTable("Validations",
		func(uac string, expected bool) {
			Expect(uacGenerator.ValidateUAC12(uac)).To(Equal(expected))
		},
		Entry("short", "21314", false),
		Entry("long", "21314632512345123", false),
		Entry("letters", "abcdabcdabcd", false),
		Entry("badnumbers", "1234012341234", false),
		Entry("goodnumbers", "123412341234", true),
	)
})

var _ = Describe("ValidateUAC16", func() {
	var uacGenerator = &uacgenerator.UacGenerator{}
	DescribeTable("Validations",
		func(uac string, expected bool) {
			Expect(uacGenerator.ValidateUAC16(uac)).To(Equal(expected))
		},
		Entry("short", "21314", false),
		Entry("long", "21314632512345123", false),
		Entry("vowles", "abcdabcdabcdabcd", false),
		Entry("ones", "1111222233334444", false),
		Entry("zeroes", "0000222233334444", false),
		Entry("all letters", "mnbvmnbvmnbvmnbv", true),
		Entry("all numbers", "2345678923456789", true),
		Entry("mix", "23kl56mn78fd42bn", true),
	)
})

var _ = Describe("ValidateUAC", func() {
	var (
		uacGenerator = &uacgenerator.UacGenerator{}
		uac12        = "123412341234"
		uac16        = "23kl56mn78fd42bn"
	)
	Context("when configured for 12 digit UACs", func() {
		BeforeEach(func() {
			uacGenerator.UacKind = "uac"
		})

		Context("when a 16 character UAC is provided", func() {
			It("returns false", func() {
				Expect(uacGenerator.ValidateUAC(uac16)).To(BeFalse())
			})
		})

		Context("when a 12 digit UAC is provided", func() {
			It("returns true", func() {
				Expect(uacGenerator.ValidateUAC(uac12)).To(BeTrue())
			})
		})
	})

	Context("when configured for 16 character UACs", func() {
		BeforeEach(func() {
			uacGenerator.UacKind = "uac16"
		})

		Context("when a 16 character UAC is provided", func() {
			It("returns true", func() {
				Expect(uacGenerator.ValidateUAC(uac16)).To(BeTrue())
			})
		})

		Context("when a 12 digit UAC is provided", func() {
			It("returns false", func() {
				Expect(uacGenerator.ValidateUAC(uac12)).To(BeFalse())
			})
		})
	})
})

var _ = Describe("ValiadateUACs", func() {
	var uacGenerator = &uacgenerator.UacGenerator{}
	Context("when some uacs are invalid", func() {
		var uacs = []string{"2313", "41512", "123412341234"}

		It("returms an ImportError with invalid UACs", func() {
			err := uacGenerator.ValidateUACs(uacs)
			Expect(err.(*uacgenerator.ImportError).InvalidUACs).To(Equal([]string{"2313", "41512"}))
		})
	})

	Context("when all uacs are valid", func() {
		var uacs = []string{"123412341234", "456745674567"}

		It("returns nil", func() {
			Expect(uacGenerator.ValidateUACs(uacs)).To(BeNil())
		})
	})
})
