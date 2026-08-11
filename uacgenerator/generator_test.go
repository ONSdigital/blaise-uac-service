package uacgenerator_test

import (
	"context"
	"errors"
	"fmt"

	"cloud.google.com/go/datastore"
	"github.com/ONSDigital/blaise-uac-service/uacgenerator"
	"github.com/ONSDigital/blaise-uac-service/uacgenerator/mocks"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ = Describe("Generate for a single case", func() {
	var (
		uacService     *uacgenerator.UACService
		instrumentName = "lolcat"
		caseID         = "74628568"
		mockDatastore  *mocks.DatastoreInterface
	)

	BeforeEach(func() {
		mockDatastore = &mocks.DatastoreInterface{}

		mockDatastore.On("GetAll",
			context.Background(),
			mock.AnythingOfType("*datastore.Query"),
			mock.AnythingOfType("*[]*uacgenerator.UACInfo"),
		).Return(nil, nil)
	})

	Context("when a UAC kind is blank", func() {
		BeforeEach(func() {
			uacService = uacgenerator.NewUACService(mockDatastore, "")
		})

		It("returns an error", func() {
			err := uacService.Generate(context.Background(), instrumentName, []string{caseID})
			Expect(errors.Is(err, uacgenerator.ErrInvalidUACKind)).To(BeTrue())
		})
	})

	Context("when a UAC kind is invalid", func() {
		BeforeEach(func() {
			uacService = uacgenerator.NewUACService(mockDatastore, "this is not a valid UWACKY")
		})

		It("returns an error", func() {
			err := uacService.Generate(context.Background(), instrumentName, []string{caseID})
			Expect(errors.Is(err, uacgenerator.ErrInvalidUACKind)).To(BeTrue())
		})
	})

	Context("when a case ID is blank", func() {
		BeforeEach(func() {
			uacService = uacgenerator.NewUACService(mockDatastore, "uac")
		})

		It("returns an error", func() {
			err := uacService.Generate(context.Background(), instrumentName, []string{""})
			Expect(errors.Is(err, uacgenerator.ErrBlankCaseID)).To(BeTrue())
		})
	})

	Context("when a generated UAC already exists in datastore", func() {
		BeforeEach(func() {
			uacService = uacgenerator.NewUACService(mockDatastore, "uac")

			mockDatastore.On("Mutate",
				context.Background(),
				mock.AnythingOfType("*datastore.Mutation"),
			).Twice().Return(nil, status.Error(codes.AlreadyExists, "Already exists"))
			mockDatastore.On("Mutate",
				context.Background(),
				mock.AnythingOfType("*datastore.Mutation"),
			).Return(nil, nil)
		})

		It("regenerates a new random UAC and saves it to datastore", func() {
			err := uacService.Generate(context.Background(), instrumentName, []string{caseID})
			Expect(err).ShouldNot(HaveOccurred())
			mockDatastore.AssertNumberOfCalls(GinkgoT(), "Mutate", 3)
		})
	})

	Context("when a generated UAC does not exist in datastore", func() {
		BeforeEach(func() {
			uacService = uacgenerator.NewUACService(mockDatastore, "uac")

			mockDatastore.On("Mutate",
				context.Background(),
				mock.AnythingOfType("*datastore.Mutation"),
			).Return(nil, nil)
		})

		It("saves the UAC to datastore", func() {
			err := uacService.Generate(context.Background(), instrumentName, []string{caseID})
			Expect(err).ShouldNot(HaveOccurred())
			mockDatastore.AssertNumberOfCalls(GinkgoT(), "Mutate", 1)
		})
	})

	Context("when a generated UAC already exists in datastore over 10 times", func() {
		BeforeEach(func() {
			uacService = uacgenerator.NewUACService(mockDatastore, "uac")

			mockDatastore.On("Mutate",
				context.Background(),
				mock.AnythingOfType("*datastore.Mutation"),
			).Return(nil, status.Error(codes.AlreadyExists, "Already exists"))
		})

		It("gives up generating a UAC and returns an error", func() {
			err := uacService.Generate(context.Background(), instrumentName, []string{caseID})
			mockDatastore.AssertNumberOfCalls(GinkgoT(), "Mutate", 10)
			Expect(errors.Is(err, uacgenerator.ErrCouldNotGenerateUniqueUAC)).To(BeTrue())
		})
	})
})

var _ = Describe("Generate", func() {
	var (
		uacService     *uacgenerator.UACService
		instrumentName = "lolcat"
		caseIDs        = []string{
			"74628568",
			"74628561",
			"74628562",
			"74628563",
			"74628564",
		}
		mockDatastore *mocks.DatastoreInterface
	)

	Context("when none of the cases have a UAC", func() {
		BeforeEach(func() {
			mockDatastore = &mocks.DatastoreInterface{}

			uacService = uacgenerator.NewUACService(mockDatastore, "uac")

			mockDatastore.On("GetAll",
				context.Background(),
				mock.AnythingOfType("*datastore.Query"),
				mock.AnythingOfType("*[]*uacgenerator.UACInfo"),
			).Return(nil, nil)

			mockDatastore.On("Mutate",
				context.Background(),
				mock.AnythingOfType("*datastore.Mutation"),
			).Return(nil, nil)
		})

		It("generates UACs for all case IDs in an instrument", func() {
			Expect(uacService.Generate(context.Background(), instrumentName, caseIDs)).To(BeNil())

			mockDatastore.AssertNumberOfCalls(GinkgoT(), "Mutate", len(caseIDs))
			mockDatastore.AssertNumberOfCalls(GinkgoT(), "GetAll", len(caseIDs))
		})
	})

	Context("when at least one generation errors", func() {
		BeforeEach(func() {
			mockDatastore = &mocks.DatastoreInterface{}

			uacService = uacgenerator.NewUACService(mockDatastore, "uac")

			mockDatastore.On("GetAll",
				context.Background(),
				mock.AnythingOfType("*datastore.Query"),
				mock.AnythingOfType("*[]*uacgenerator.UACInfo"),
			).Return(nil, nil)

			mockDatastore.On("Mutate",
				context.Background(),
				mock.AnythingOfType("*datastore.Mutation"),
			).Once().Return(nil, nil)
			mockDatastore.On("Mutate",
				context.Background(),
				mock.AnythingOfType("*datastore.Mutation"),
			).Once().Return(nil, fmt.Errorf("Massive mutation explosion"))
			mockDatastore.On("Mutate",
				context.Background(),
				mock.AnythingOfType("*datastore.Mutation"),
			).Return(nil, nil)
		})

		It("returns an error", func() {
			Expect(uacService.Generate(context.Background(), instrumentName, caseIDs)).To(MatchError("Massive mutation explosion"))
		})
	})

	Context("when one of the cases already has a UAC", func() {
		BeforeEach(func() {
			mockDatastore = &mocks.DatastoreInterface{}

			uacService = uacgenerator.NewUACService(mockDatastore, "uac")

			mockDatastore.On("GetAll",
				context.Background(),
				mock.AnythingOfType("*datastore.Query"),
				mock.AnythingOfType("*[]*uacgenerator.UACInfo"),
			).Once().Return([]*datastore.Key{datastore.IncompleteKey("foo", nil)}, nil)

			mockDatastore.On("GetAll",
				context.Background(),
				mock.AnythingOfType("*datastore.Query"),
				mock.AnythingOfType("*[]*uacgenerator.UACInfo"),
			).Return(nil, nil)

			mockDatastore.On("Mutate",
				context.Background(),
				mock.AnythingOfType("*datastore.Mutation"),
			).Return(nil, nil)
		})

		It("generates UACs for all case IDs in an instrument", func() {
			Expect(uacService.Generate(context.Background(), instrumentName, caseIDs)).To(BeNil())

			mockDatastore.AssertNumberOfCalls(GinkgoT(), "Mutate", len(caseIDs)-1)
			mockDatastore.AssertNumberOfCalls(GinkgoT(), "GetAll", len(caseIDs))
		})
	})

	Context("when there are no cases", func() {
		BeforeEach(func() {
			mockDatastore = &mocks.DatastoreInterface{}

			uacService = uacgenerator.NewUACService(mockDatastore, "uac")

			mockDatastore.On("GetAll",
				context.Background(),
				mock.AnythingOfType("*datastore.Query"),
				mock.AnythingOfType("*[]*uacgenerator.UACInfo"),
			).Return(nil, nil)

			mockDatastore.On("Mutate",
				context.Background(),
				mock.AnythingOfType("*datastore.Mutation"),
			).Return(nil, nil)
		})

		It("generates UACs for all case IDs in an instrument", func() {
			Expect(uacService.Generate(context.Background(), instrumentName, []string{})).To(BeNil())

			mockDatastore.AssertNumberOfCalls(GinkgoT(), "Mutate", 0)
			mockDatastore.AssertNumberOfCalls(GinkgoT(), "GetAll", 0)
		})
	})
})

var _ = Describe("GetAllUACs", func() {
	var (
		uacService     *uacgenerator.UACService
		instrumentName = "lolcat"
		mockDatastore  *mocks.DatastoreInterface
	)

	BeforeEach(func() {
		mockDatastore = &mocks.DatastoreInterface{}

		uacService = uacgenerator.NewUACService(mockDatastore, "uac")

		mockDatastore.On("GetAll",
			context.Background(),
			mock.AnythingOfType("*datastore.Query"),
			mock.AnythingOfType("*[]*uacgenerator.UACInfo"),
		).Once().Return(
			func(ctx context.Context, qry *datastore.Query, dst interface{}) []*datastore.Key {
				uacInfos := dst.(*[]*uacgenerator.UACInfo)
				key := datastore.NameKey(uacService.UACKind, "foobar", nil)
				*uacInfos = append(*uacInfos, &uacgenerator.UACInfo{
					InstrumentName: instrumentName,
					CaseID:         "12343",
					UAC:            key,
				})
				key2 := datastore.NameKey(uacService.UACKind, "foobar2", nil)
				*uacInfos = append(*uacInfos, &uacgenerator.UACInfo{
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

	It("returns a map of all UACs with info", func() {
		uacs, err := uacService.GetAllUACs(context.Background(), instrumentName)
		Expect(uacs).To(HaveLen(2))
		Expect(uacs["foobar"].InstrumentName).To(Equal(instrumentName))
		Expect(uacs["foobar"].CaseID).To(Equal("12343"))
		Expect(uacs["foobar2"].InstrumentName).To(Equal(instrumentName))
		Expect(uacs["foobar2"].CaseID).To(Equal("56764"))
		Expect(err).To(BeNil())
	})
})

var _ = Describe("GetAllUACsByCaseID", func() {
	var (
		uacService     *uacgenerator.UACService
		instrumentName = "lolcat"
		mockDatastore  *mocks.DatastoreInterface
	)

	Context("when there are duplicate case IDs", func() {
		BeforeEach(func() {
			mockDatastore = &mocks.DatastoreInterface{}

			uacService = uacgenerator.NewUACService(mockDatastore, "uac")

			mockDatastore.On("GetAll",
				context.Background(),
				mock.AnythingOfType("*datastore.Query"),
				mock.AnythingOfType("*[]*uacgenerator.UACInfo"),
			).Once().Return(
				func(ctx context.Context, qry *datastore.Query, dst interface{}) []*datastore.Key {
					uacInfos := dst.(*[]*uacgenerator.UACInfo)
					key := datastore.NameKey(uacService.UACKind, "foobar", nil)
					*uacInfos = append(*uacInfos, &uacgenerator.UACInfo{
						InstrumentName: instrumentName,
						CaseID:         "12343",
						UAC:            key,
					})
					key2 := datastore.NameKey(uacService.UACKind, "foobar2", nil)
					*uacInfos = append(*uacInfos, &uacgenerator.UACInfo{
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
			uacs, err := uacService.GetAllUACsByCaseID(context.Background(), instrumentName)
			Expect(uacs).To(BeNil())
			Expect(err).To(MatchError("fewer case IDs than UACs: duplicate case IDs detected"))
		})
	})

	Context("when there are no duplicate case IDs", func() {
		BeforeEach(func() {
			mockDatastore = &mocks.DatastoreInterface{}

			uacService = uacgenerator.NewUACService(mockDatastore, "uac")

			mockDatastore.On("GetAll",
				context.Background(),
				mock.AnythingOfType("*datastore.Query"),
				mock.AnythingOfType("*[]*uacgenerator.UACInfo"),
			).Once().Return(
				func(ctx context.Context, qry *datastore.Query, dst interface{}) []*datastore.Key {
					uacInfos := dst.(*[]*uacgenerator.UACInfo)
					key := datastore.NameKey(uacService.UACKind, "foobar", nil)
					*uacInfos = append(*uacInfos, &uacgenerator.UACInfo{
						InstrumentName: instrumentName,
						CaseID:         "12343",
						UAC:            key,
					})
					key2 := datastore.NameKey(uacService.UACKind, "foobar2", nil)
					*uacInfos = append(*uacInfos, &uacgenerator.UACInfo{
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

		It("returns a map of all UACs with info", func() {
			uacs, err := uacService.GetAllUACsByCaseID(context.Background(), instrumentName)
			Expect(uacs).To(HaveLen(2))
			Expect(uacs["12343"].InstrumentName).To(Equal(instrumentName))
			Expect(uacs["12343"].CaseID).To(Equal("12343"))
			Expect(uacs["56764"].InstrumentName).To(Equal(instrumentName))
			Expect(uacs["56764"].CaseID).To(Equal("56764"))
			Expect(err).To(BeNil())
		})
	})
})

var _ = Describe("GetUACCount", func() {
	var (
		uacService     *uacgenerator.UACService
		instrumentName = "lolcat"
		mockDatastore  *mocks.DatastoreInterface
	)

	BeforeEach(func() {
		mockDatastore = &mocks.DatastoreInterface{}

		uacService = uacgenerator.NewUACService(mockDatastore, "uac")

		mockDatastore.On("Count",
			context.Background(),
			mock.AnythingOfType("*datastore.Query"),
		).Return(40, nil)
	})

	It("returns the UAC count", func() {
		count, err := uacService.GetUACCount(context.Background(), instrumentName)
		Expect(count).To(Equal(40))
		Expect(err).To(BeNil())
	})
})

var _ = Describe("GetUACInfo", func() {
	var (
		uacService     *uacgenerator.UACService
		instrumentName = "lolcat"
		mockDatastore  *mocks.DatastoreInterface
	)

	BeforeEach(func() {
		mockDatastore = &mocks.DatastoreInterface{}

		uacService = uacgenerator.NewUACService(mockDatastore, "uac")

		mockDatastore.On("Get",
			context.Background(),
			mock.AnythingOfType("*datastore.Key"),
			mock.AnythingOfType("*uacgenerator.UACInfo"),
		).Once().Return(
			func(ctx context.Context, keyQry *datastore.Key, dst interface{}) error {
				uacInfo := dst.(*uacgenerator.UACInfo)
				key := datastore.NameKey(uacService.UACKind, "lemons", nil)
				*uacInfo = uacgenerator.UACInfo{
					InstrumentName: instrumentName,
					CaseID:         "12343",
					UAC:            key,
				}
				return nil
			})
	})

	It("returns UAC info for a valid UAC", func() {
		uacInfo, err := uacService.GetUACInfo(context.Background(), "lemons")
		Expect(uacInfo.InstrumentName).To(Equal(instrumentName))
		Expect(uacInfo.CaseID).To(Equal("12343"))
		Expect(err).To(BeNil())
	})
})

var _ = Describe("GetInstruments", func() {
	var (
		uacService    *uacgenerator.UACService
		mockDatastore *mocks.DatastoreInterface
	)

	BeforeEach(func() {
		mockDatastore = &mocks.DatastoreInterface{}

		uacService = uacgenerator.NewUACService(mockDatastore, "uac")

		mockDatastore.On("GetAll",
			context.Background(),
			mock.AnythingOfType("*datastore.Query"),
			mock.AnythingOfType("*[]*uacgenerator.UACInfo"),
		).Once().Return(
			func(ctx context.Context, qry *datastore.Query, dst interface{}) []*datastore.Key {
				uacInfos := dst.(*[]*uacgenerator.UACInfo)
				*uacInfos = append(*uacInfos, &uacgenerator.UACInfo{
					InstrumentName: "foo",
				})
				*uacInfos = append(*uacInfos, &uacgenerator.UACInfo{
					InstrumentName: "bar",
				})
				return []*datastore.Key{}
			},
			func(ctx context.Context, qry *datastore.Query, dst interface{}) error {
				return nil
			})
	})

	It("returns a list of instrument names", func() {
		instrumentNames, err := uacService.GetInstruments(context.Background())
		Expect(err).To(BeNil())
		Expect(instrumentNames).To(Equal([]string{"foo", "bar"}))
	})
})

var _ = Describe("UACs", func() {
	Describe("BuildUACChunks", func() {
		var uacs = uacgenerator.UACs{
			"111122223333": &uacgenerator.UACInfo{},
			"123456781234": &uacgenerator.UACInfo{},
		}

		It("adds UAC chunks to UAC info", func() {
			uacs.BuildUACChunks()
			Expect(*uacs["111122223333"].UACChunks).To(Equal(uacgenerator.UACChunks{UAC1: "1111", UAC2: "2222", UAC3: "3333"}))
			Expect(*uacs["123456781234"].UACChunks).To(Equal(uacgenerator.UACChunks{UAC1: "1234", UAC2: "5678", UAC3: "1234"}))
		})
	})
})

var _ = Describe("ImportUACs", func() {
	var (
		uacService    *uacgenerator.UACService
		mockDatastore *mocks.DatastoreInterface
		uacs          []string
	)

	BeforeEach(func() {
		mockDatastore = &mocks.DatastoreInterface{}
		uacService = uacgenerator.NewUACService(mockDatastore, "uac")

		mockDatastore.On("Mutate",
			context.Background(),
			mock.AnythingOfType("*datastore.Mutation"),
		).Return(nil, nil)
	})

	AfterEach(func() {
		uacs = []string{}
	})

	Context("when there are no UACs", func() {
		BeforeEach(func() {
			uacs = []string{}
		})

		It("imports nothing and returns 0 imported with no error", func() {
			updateCount, err := uacService.ImportUACs(context.Background(), uacs)
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
					context.Background(),
					mock.AnythingOfType("*datastore.Key"),
					mock.AnythingOfType("*uacgenerator.UACInfo"),
				).Return(datastore.ErrNoSuchEntity)
			})

			It("imports all of the UACs", func() {
				updateCount, err := uacService.ImportUACs(context.Background(), uacs)
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
				updateCount, err := uacService.ImportUACs(context.Background(), uacs)
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
				context.Background(),
				mock.AnythingOfType("*datastore.Key"),
				mock.AnythingOfType("*uacgenerator.UACInfo"),
			).Return(func(ctx context.Context, keyQry *datastore.Key, dst interface{}) error {
				uacInfo := dst.(*uacgenerator.UACInfo)
				key := datastore.NameKey(uacService.UACKind, "any", nil)
				*uacInfo = uacgenerator.UACInfo{
					InstrumentName: "unknown",
					CaseID:         "unknown",
					UAC:            key,
				}
				return nil
			})
		})

		It("imports nothing and returns 0 imported with no error", func() {
			updateCount, err := uacService.ImportUACs(context.Background(), uacs)
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
					context.Background(),
					mock.AnythingOfType("*datastore.Key"),
					mock.AnythingOfType("*uacgenerator.UACInfo"),
				).Times(2).Return(datastore.ErrNoSuchEntity)
				mockDatastore.On("Get",
					context.Background(),
					mock.AnythingOfType("*datastore.Key"),
					mock.AnythingOfType("*uacgenerator.UACInfo"),
				).Return(func(ctx context.Context, keyQry *datastore.Key, dst interface{}) error {
					uacInfo := dst.(*uacgenerator.UACInfo)
					key := datastore.NameKey(uacService.UACKind, "123556789987", nil)
					*uacInfo = uacgenerator.UACInfo{
						InstrumentName: "unknown",
						CaseID:         "unknown",
						UAC:            key,
					}
					return nil
				})
			})

			It("imports all of the UACs, skipping those that already exist", func() {
				updateCount, err := uacService.ImportUACs(context.Background(), uacs)
				Expect(updateCount).To(Equal(2))
				Expect(err).To(BeNil())
				mockDatastore.AssertNumberOfCalls(GinkgoT(), "Mutate", 2)
			})
		})

		Context("and they have InstrumentNames that are not 'unknown'", func() {
			BeforeEach(func() {
				mockDatastore.On("Get",
					context.Background(),
					datastore.NameKey(uacService.UACKind, "123556789987", nil),
					mock.AnythingOfType("*uacgenerator.UACInfo"),
				).Return(func(ctx context.Context, keyQry *datastore.Key, dst interface{}) error {
					uacInfo := dst.(*uacgenerator.UACInfo)
					key := datastore.NameKey(uacService.UACKind, "123556789987", nil)
					*uacInfo = uacgenerator.UACInfo{
						InstrumentName: "dst2108a",
						CaseID:         "1234",
						UAC:            key,
					}
					return nil
				})
				mockDatastore.On("Get",
					context.Background(),
					mock.AnythingOfType("*datastore.Key"),
					mock.AnythingOfType("*uacgenerator.UACInfo"),
				).Return(datastore.ErrNoSuchEntity)
			})

			It("errors and doesn't import anything", func() {
				updateCount, err := uacService.ImportUACs(context.Background(), uacs)
				Expect(updateCount).To(Equal(0))
				Expect(err).To(MatchError(`Cannot import UACs because some were already in use by questionnaires: ["123556789987"]`))
				mockDatastore.AssertNumberOfCalls(GinkgoT(), "Mutate", 0)
			})
		})
	})
})

var _ = Describe("GetAllDisabledUACs", func() {
	var (
		uacService     *uacgenerator.UACService
		instrumentName = "lolcat"
		mockDatastore  *mocks.DatastoreInterface
	)

	Context("when there are duplicate disabled UACs with the same case ID", func() {
		BeforeEach(func() {
			mockDatastore = &mocks.DatastoreInterface{}

			uacService = uacgenerator.NewUACService(mockDatastore, "uac")

			mockDatastore.On("GetAll",
				context.Background(),
				mock.AnythingOfType("*datastore.Query"),
				mock.AnythingOfType("*[]*uacgenerator.UACInfo"),
			).Once().Return(
				func(ctx context.Context, qry *datastore.Query, dst interface{}) []*datastore.Key {
					uacInfos := dst.(*[]*uacgenerator.UACInfo)
					key := datastore.NameKey(uacService.UACKind, "foobar", nil)
					*uacInfos = append(*uacInfos, &uacgenerator.UACInfo{
						InstrumentName: instrumentName,
						CaseID:         "12343",
						UAC:            key,
						Disabled:       true,
					})
					key2 := datastore.NameKey(uacService.UACKind, "foobar2", nil)
					*uacInfos = append(*uacInfos, &uacgenerator.UACInfo{
						InstrumentName: instrumentName,
						CaseID:         "12343",
						UAC:            key2,
						Disabled:       true,
					})
					return []*datastore.Key{key, key2}
				},
				func(ctx context.Context, qry *datastore.Query, dst interface{}) error {
					return nil
				})
		})

		It("returns an error", func() {
			uacs, err := uacService.GetAllDisabledUACs(context.Background(), instrumentName)
			Expect(uacs).To(BeNil())
			Expect(err).To(MatchError("fewer case IDs than UACs: duplicate case IDs detected"))
		})
	})

	Context("when there are no duplicate case IDs", func() {
		BeforeEach(func() {
			mockDatastore = &mocks.DatastoreInterface{}

			uacService = uacgenerator.NewUACService(mockDatastore, "uac")

			mockDatastore.On("GetAll",
				context.Background(),
				mock.AnythingOfType("*datastore.Query"),
				mock.AnythingOfType("*[]*uacgenerator.UACInfo"),
			).Once().Return(
				func(ctx context.Context, qry *datastore.Query, dst interface{}) []*datastore.Key {
					uacInfos := dst.(*[]*uacgenerator.UACInfo)
					key := datastore.NameKey(uacService.UACKind, "foobar", nil)
					*uacInfos = append(*uacInfos, &uacgenerator.UACInfo{
						InstrumentName: instrumentName,
						CaseID:         "12343",
						UAC:            key,
						Disabled:       true,
					})
					key2 := datastore.NameKey(uacService.UACKind, "foobar2", nil)
					*uacInfos = append(*uacInfos, &uacgenerator.UACInfo{
						InstrumentName: instrumentName,
						CaseID:         "56764",
						UAC:            key2,
						Disabled:       true,
					})
					return []*datastore.Key{key, key2}
				},
				func(ctx context.Context, qry *datastore.Query, dst interface{}) error {
					return nil
				})
		})

		It("returns a map of all UACs with info", func() {
			uacs, err := uacService.GetAllDisabledUACs(context.Background(), instrumentName)
			Expect(uacs).To(HaveLen(2))
			Expect(uacs["12343"].InstrumentName).To(Equal(instrumentName))
			Expect(uacs["12343"].CaseID).To(Equal("12343"))
			Expect(uacs["12343"].Disabled).To(Equal(true))
			Expect(uacs["56764"].InstrumentName).To(Equal(instrumentName))
			Expect(uacs["56764"].CaseID).To(Equal("56764"))
			Expect(uacs["56764"].Disabled).To(Equal(true))
			Expect(err).To(BeNil())
		})
	})
})

var _ = Describe("EnableUAC", func() {
	var (
		uacService    *uacgenerator.UACService
		mockDatastore *mocks.DatastoreInterface
		uac           string
	)

	BeforeEach(func() {
		mockDatastore = &mocks.DatastoreInterface{}
		uacService = uacgenerator.NewUACService(mockDatastore, "uac")

		mockDatastore.On("Mutate",
			context.Background(),
			mock.AnythingOfType("*datastore.Mutation"),
		).Return(nil, nil)
	})

	AfterEach(func() {
		uac = ""
	})

	Context("and the UAC is enabled", func() {
		BeforeEach(func() {
			uac = "123456789123"
		})

		Context("and they have a Disabled attribute of false", func() {
			BeforeEach(func() {
				mockDatastore.On("Get",
					context.Background(),
					mock.AnythingOfType("*datastore.Key"),
					mock.AnythingOfType("*uacgenerator.UACInfo"),
				).Times(1).Return(func(ctx context.Context, keyQry *datastore.Key, dst interface{}) error {
					uacInfo := dst.(*uacgenerator.UACInfo)
					key := datastore.NameKey(uacService.UACKind, "123456789123", nil)
					*uacInfo = uacgenerator.UACInfo{
						InstrumentName: "dst2108a",
						CaseID:         "1234",
						UAC:            key,
						Disabled:       false,
					}
					return nil
				})
			})

			It("enables the UAC", func() {
				err := uacService.EnableUAC(context.Background(), uac)
				Expect(err).To(BeNil())
				mockDatastore.AssertNumberOfCalls(GinkgoT(), "Mutate", 1)
			})
		})
	})

	Context("and the UAC doesn't exist", func() {
		BeforeEach(func() {
			uac = "123456789123"
		})

		Context("and should return an error", func() {
			BeforeEach(func() {
				mockDatastore.On("Get",
					context.Background(),
					mock.AnythingOfType("*datastore.Key"),
					mock.AnythingOfType("*uacgenerator.UACInfo"),
				).Return(datastore.ErrNoSuchEntity)
			})

			It("errors and doesn't enable anything", func() {
				err := uacService.EnableUAC(context.Background(), uac)
				Expect(errors.Is(err, uacgenerator.ErrInvalidUAC)).To(BeTrue())
				mockDatastore.AssertNumberOfCalls(GinkgoT(), "Mutate", 0)
			})
		})
	})
})

var _ = Describe("DisableUAC", func() {
	var (
		uacService    *uacgenerator.UACService
		mockDatastore *mocks.DatastoreInterface
		uac           string
	)

	BeforeEach(func() {
		mockDatastore = &mocks.DatastoreInterface{}
		uacService = uacgenerator.NewUACService(mockDatastore, "uac")

		mockDatastore.On("Mutate",
			context.Background(),
			mock.AnythingOfType("*datastore.Mutation"),
		).Return(nil, nil)
	})

	AfterEach(func() {
		uac = ""
	})

	Context("and the UAC is disabled", func() {
		BeforeEach(func() {
			uac = "123456789123"
		})

		Context("and they have a Disabled attribute of true", func() {
			BeforeEach(func() {
				mockDatastore.On("Get",
					context.Background(),
					mock.AnythingOfType("*datastore.Key"),
					mock.AnythingOfType("*uacgenerator.UACInfo"),
				).Times(1).Return(func(ctx context.Context, keyQry *datastore.Key, dst interface{}) error {
					uacInfo := dst.(*uacgenerator.UACInfo)
					key := datastore.NameKey(uacService.UACKind, "123456789123", nil)
					*uacInfo = uacgenerator.UACInfo{
						InstrumentName: "dst2108a",
						CaseID:         "1234",
						UAC:            key,
						Disabled:       true,
					}
					return nil
				})
			})

			It("disables the UAC", func() {
				err := uacService.DisableUAC(context.Background(), uac)
				Expect(err).To(BeNil())
				mockDatastore.AssertNumberOfCalls(GinkgoT(), "Mutate", 1)
			})
		})
	})

	Context("and the UAC doesn't exist", func() {
		BeforeEach(func() {
			uac = "123456789123"
		})

		Context("and should return an error", func() {
			BeforeEach(func() {
				mockDatastore.On("Get",
					context.Background(),
					mock.AnythingOfType("*datastore.Key"),
					mock.AnythingOfType("*uacgenerator.UACInfo"),
				).Return(datastore.ErrNoSuchEntity)
			})

			It("errors and doesn't disable anything", func() {
				err := uacService.DisableUAC(context.Background(), uac)
				Expect(errors.Is(err, uacgenerator.ErrInvalidUAC)).To(BeTrue())
				mockDatastore.AssertNumberOfCalls(GinkgoT(), "Mutate", 0)
			})
		})
	})
})
