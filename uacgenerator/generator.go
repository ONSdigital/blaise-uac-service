package uacgenerator

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"cloud.google.com/go/datastore"
	"github.com/zenthangplus/goccm"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	MAXCONCURRENT      = 500
	APPROVEDCHARACTERS = "bcdfghjklmnpqrstvxz23456789"
	UNKNOWNINSTRUMENT  = "unknown"
)

// Generate mocks by running "go generate ./..."
//
//go:generate mockery --name UacGeneratorInterface
type UacGeneratorInterface interface {
	Generate(string, []string) error
	GetAllUacs(string) (Uacs, error)
	GetAllUacsByCaseID(string) (Uacs, error)
	GetAllUacsDisabled(string) (Uacs, error)
	GetUacCount(string) (int, error)
	GetUacInfo(string) (*UacInfo, error)
	GetInstruments() ([]string, error)
	ImportUACs([]string) (int, error)
	AdminDelete(string) error
	DisableUac(string) error
	EnableUac(string) error
}

// Generate mocks by running "go generate ./..."
//
//go:generate mockery --name Datastore
type Datastore interface {
	Mutate(context.Context, ...*datastore.Mutation) ([]*datastore.Key, error)
	GetAll(context.Context, *datastore.Query, interface{}) ([]*datastore.Key, error)
	Count(context.Context, *datastore.Query) (int, error)
	Get(context.Context, *datastore.Key, interface{}) error
	DeleteMulti(context.Context, []*datastore.Key) error
	Close() error
}

type UacChunks struct {
	UAC1 string `json:"uac1"`
	UAC2 string `json:"uac2"`
	UAC3 string `json:"uac3"`
	UAC4 string `json:"uac4,omitempty"`
}

type UacGenerator struct {
	UacKind         string
	DatastoreClient Datastore
	Context         context.Context
	GenerateError   map[string]error
	Randomizer      *rand.Rand
	mu              sync.Mutex
	importMu        sync.Mutex
}

type UacInfo struct {
	InstrumentName string         `json:"instrument_name" datastore:"instrument_name"`
	CaseID         string         `json:"case_id" datastore:"case_id"`
	UacChunks      *UacChunks     `json:"uac_chunks,omitempty" datastore:"-"`
	UAC            *datastore.Key `json:"-" datastore:"__key__"`
	FullUAC        string         `json:"full_uac,omitempty" datastore:"-"`
	Disabled       bool           `json:"disabled" datastore:"disabled"`
}

type Uacs map[string]*UacInfo

func (uacs Uacs) BuildUacChunks() {
	for uac, uacInfo := range uacs {
		if uacInfo.FullUAC != "" {
			uac = uacInfo.FullUAC
		}
		uacInfo.UacChunks = ChunkUAC(uac)
	}
}

func NewUacGenerator(datastoreClient Datastore, uacKind string) *UacGenerator {
	return &UacGenerator{
		UacKind:         uacKind,
		Context:         context.Background(),
		Randomizer:      rand.New(cryptoSource{}),
		DatastoreClient: datastoreClient,
	}
}

func (uacGenerator *UacGenerator) GenerateUac12() string {
	var uac string
	for i := 0; i < 3; i++ {
		uacSegmant := uacGenerator.Randomizer.Int63n(9999 - 1000)
		uac = fmt.Sprintf("%s%d", uac, uacSegmant+1000)
	}
	return uac
}

func (uacGenerator *UacGenerator) GenerateUac16() string {
	b := make([]byte, 16)
	for i := range b {
		b[i] = APPROVEDCHARACTERS[uacGenerator.Randomizer.Intn(len(APPROVEDCHARACTERS))]
	}
	return string(b)
}

func (uacGenerator *UacGenerator) NewUac(instrumentName, caseID string, attempt int) (string, error) {
	if caseID == "" {
		return "", fmt.Errorf("Cannot generate UACs for blank caseIDs")
	}
	if attempt >= 10 {
		return "", fmt.Errorf("Could not generate a unique UAC in 10 attempts")
	}

	var uac string
	switch uacGenerator.UacKind {
	case "uac":
		uac = uacGenerator.GenerateUac12()
	case "uac16":
		uac = uacGenerator.GenerateUac16()
	default:
		return "", fmt.Errorf("Cannot generate UACs for invalid UacKind")
	}

	err := uacGenerator.AddUacToDatastore(uac, instrumentName, caseID)
	if err != nil {
		if alreadyExistsError(err) {
			return uacGenerator.NewUac(instrumentName, caseID, attempt+1)
		}
		return "", err
	}
	return uac, nil
}

func (uacGenerator *UacGenerator) AddUacToDatastore(uac string, instrumentName, caseID string) error {
	// Cannot workout how the hell to mock/ test this :(
	newUACMutation := datastore.NewInsert(uacGenerator.UacKey(uac), &UacInfo{
		InstrumentName: strings.ToLower(instrumentName),
		CaseID:         strings.ToLower(caseID),
	})
	_, err := uacGenerator.DatastoreClient.Mutate(uacGenerator.Context, newUACMutation)
	if err != nil {
		return err
	}
	return nil
}

func (uacGenerator *UacGenerator) UacKey(key string) *datastore.Key {
	return datastore.NameKey(uacGenerator.UacKind, key, nil)
}

func (uacGenerator *UacGenerator) UacExistsForCase(instrumentName, caseID string) (bool, error) {
	var existingUACs []*UacInfo
	existingUACKeys, err := uacGenerator.DatastoreClient.GetAll(
		uacGenerator.Context,
		uacGenerator.instrumentCaseQuery(instrumentName, caseID),
		&existingUACs,
	)
	if err != nil {
		return false, err
	}
	if len(existingUACKeys) >= 1 {
		return true, nil
	}
	return false, nil
}

func (uacGenerator *UacGenerator) GenerateUniqueUac(instrumentName, caseID string) error {
	exists, err := uacGenerator.UacExistsForCase(instrumentName, caseID)
	if err != nil {
		uacGenerator.mu.Lock()
		uacGenerator.GenerateError[instrumentName] = err
		uacGenerator.mu.Unlock()
		log.Println(err)
		return err
	}
	if !exists {
		_, err := uacGenerator.NewUac(instrumentName, caseID, 0)
		if err != nil {
			uacGenerator.mu.Lock()
			uacGenerator.GenerateError[instrumentName] = err
			uacGenerator.mu.Unlock()
			log.Println(err)
			return err
		}
	}
	return nil
}

func (uacGenerator *UacGenerator) Generate(instrumentName string, caseIDs []string) error {
	if len(caseIDs) == 0 {
		return nil
	}
	if uacGenerator.GenerateError == nil {
		uacGenerator.GenerateError = make(map[string]error)
	}
	concurrent := goccm.New(MAXCONCURRENT)
	for _, caseID := range caseIDs {
		concurrent.Wait()
		go func(caseID string) {
			defer concurrent.Done()
			err := uacGenerator.GenerateUniqueUac(instrumentName, caseID)
			if err != nil {
				uacGenerator.mu.Lock()
				uacGenerator.GenerateError[instrumentName] = err
				uacGenerator.mu.Unlock()
			}
		}(caseID)
	}
	concurrent.WaitAllDone()
	err := uacGenerator.GenerateError[instrumentName]
	uacGenerator.mu.Lock()
	uacGenerator.GenerateError[instrumentName] = nil
	uacGenerator.mu.Unlock()
	return err
}

func (uacGenerator *UacGenerator) GetAllUacs(instrumentName string) (Uacs, error) {
	var uacInfos []*UacInfo
	_, err := uacGenerator.DatastoreClient.GetAll(uacGenerator.Context, uacGenerator.instrumentQuery(instrumentName), &uacInfos)
	if err != nil {
		return nil, err
	}
	uacs := make(Uacs)
	for _, uacInfo := range uacInfos {
		uacs[uacInfo.UAC.Name] = uacInfo
	}
	return uacs, nil
}

func (uacGenerator *UacGenerator) GetAllUacsByCaseID(instrumentName string) (Uacs, error) {
	var uacInfos []*UacInfo
	_, err := uacGenerator.DatastoreClient.GetAll(uacGenerator.Context, uacGenerator.instrumentQuery(instrumentName), &uacInfos)
	if err != nil {
		return nil, err
	}
	uacs := make(Uacs)
	for _, uacInfo := range uacInfos {
		uacInfo.FullUAC = uacInfo.UAC.Name
		uacs[uacInfo.CaseID] = uacInfo
	}
	if len(uacs) != len(uacInfos) {
		return nil, fmt.Errorf("Fewer case ids than uacs, must be duplicate case ids")
	}
	return uacs, nil
}

func (uacGenerator *UacGenerator) GetAllUacsDisabled(instrumentName string) (Uacs, error) {
	var uacInfos []*UacInfo
	_, err := uacGenerator.DatastoreClient.GetAll(uacGenerator.Context, uacGenerator.instrumentUacDisabledQuery(instrumentName), &uacInfos)
	if err != nil {
		return nil, err
	}
	uacs := make(Uacs)
	for _, uacInfo := range uacInfos {
		uacInfo.FullUAC = uacInfo.UAC.Name
		uacs[uacInfo.CaseID] = uacInfo
	}
	if len(uacs) != len(uacInfos) {
		return nil, fmt.Errorf("Fewer case ids than uacs, must be duplicate case ids")
	}
	return uacs, nil
}

func (uacGenerator *UacGenerator) DisableUac(uac string) error {
	uacInfo := &UacInfo{}
	err := uacGenerator.DatastoreClient.Get(uacGenerator.Context, uacGenerator.UacKey(uac), uacInfo)
	if err != nil {
		return err
	}
	newUACMutation := datastore.NewUpdate(uacGenerator.UacKey(uac), &UacInfo{
		InstrumentName: strings.ToLower(uacInfo.InstrumentName),
		CaseID:         strings.ToLower(uacInfo.CaseID),
		Disabled:       true,
	})
	_, err = uacGenerator.DatastoreClient.Mutate(uacGenerator.Context, newUACMutation)
	if err != nil {
		return err
	}
	return nil
}

func (uacGenerator *UacGenerator) EnableUac(uac string) error {
	uacInfo := &UacInfo{}
	err := uacGenerator.DatastoreClient.Get(uacGenerator.Context, uacGenerator.UacKey(uac), uacInfo)
	if err != nil {
		return err
	}
	newUACMutation := datastore.NewUpdate(uacGenerator.UacKey(uac), &UacInfo{
		InstrumentName: strings.ToLower(uacInfo.InstrumentName),
		CaseID:         strings.ToLower(uacInfo.CaseID),
		Disabled:       false,
	})
	_, err = uacGenerator.DatastoreClient.Mutate(uacGenerator.Context, newUACMutation)
	if err != nil {
		return err
	}
	return nil
}

func (uacGenerator *UacGenerator) GetUacCount(instrumentName string) (int, error) {
	return uacGenerator.DatastoreClient.Count(uacGenerator.Context, uacGenerator.instrumentQuery(instrumentName))
}

func (uacGenerator *UacGenerator) GetUacInfo(uac string) (*UacInfo, error) {
	uacInfo := &UacInfo{}
	err := uacGenerator.DatastoreClient.Get(uacGenerator.Context, uacGenerator.UacKey(uac), uacInfo)
	if err != nil {
		return nil, err
	}
	return uacInfo, nil
}

func (uacGenerator *UacGenerator) GetInstruments() ([]string, error) {
	var (
		uacInfos        []*UacInfo
		instrumentNames []string
	)
	_, err := uacGenerator.DatastoreClient.GetAll(uacGenerator.Context, uacGenerator.instrumentNamesQuery(), &uacInfos)
	if err != nil {
		return nil, err
	}
	for _, uacInfo := range uacInfos {
		instrumentNames = append(instrumentNames, uacInfo.InstrumentName)
	}
	return instrumentNames, nil
}

func (uacGenerator *UacGenerator) ImportUACs(uacs []string) (int, error) {
	if err := uacGenerator.ValidateUACs(uacs); err != nil {
		return 0, err
	}
	uacsToImport, err := uacGenerator.getUACsToImport(uacs)
	if err != nil {
		return 0, err
	}
	return uacGenerator.importUACs(uacsToImport)
}

func (uacGenerator *UacGenerator) ValidateUAC12(uac string) bool {
	if len(uac) != 12 {
		return false
	}
	chunkedUAC := ChunkUAC(uac)
	uacParts := []string{chunkedUAC.UAC1, chunkedUAC.UAC2, chunkedUAC.UAC3}
	for _, uacPart := range uacParts {
		uacInt, err := strconv.Atoi(uacPart)
		if err != nil {
			return false
		}
		if uacInt < 1000 || uacInt > 9999 {
			return false
		}
	}
	return true
}

func (uacGenerator *UacGenerator) ValidateUAC16(uac string) bool {
	uac16Regex := regexp.MustCompile(fmt.Sprintf(`^[%s]{16}$`, APPROVEDCHARACTERS))
	return uac16Regex.MatchString(uac)
}

func (uacGenerator *UacGenerator) ValidateUAC(uac string) bool {
	if uacGenerator.UacKind == "uac16" {
		return uacGenerator.ValidateUAC16(uac)
	}
	return uacGenerator.ValidateUAC12(uac)
}

func (uacGenerator *UacGenerator) ValidateUACs(uacs []string) error {
	var importError ImportError
	for _, uac := range uacs {
		if !uacGenerator.ValidateUAC(uac) {
			importError.InvalidUACs = append(importError.InvalidUACs, uac)
		}
	}
	if importError.HasErrors() {
		return &importError
	}
	return nil
}

func (uacGenerator *UacGenerator) AdminDelete(instrumentName string) error {
	instrumentUACKeys, err := uacGenerator.DatastoreClient.GetAll(uacGenerator.Context, uacGenerator.instrumentQuery(instrumentName).KeysOnly(), nil)
	if err != nil {
		return err
	}
	if len(instrumentUACKeys) == 0 {
		return nil
	}
	uacKeyChunks := chunkDatastoreKeys(instrumentUACKeys)
	concurrent := goccm.New(MAXCONCURRENT)
	for _, uacKeyChunk := range uacKeyChunks {
		concurrent.Wait()
		go func(uacKeyChunk []*datastore.Key) {
			uacGenerator.adminDeleteChunk(uacKeyChunk, concurrent)
		}(uacKeyChunk)
	}
	concurrent.WaitAllDone()
	return nil
}

func (uacGenerator *UacGenerator) getUACsToImport(uacs []string) ([]string, error) {
	var (
		uacsToImport []string
		importError  ImportError
		errors       []error
	)

	if len(uacs) == 0 {
		return nil, nil
	}

	concurrent := goccm.New(MAXCONCURRENT)
	for _, uac := range uacs {
		concurrent.Wait()
		go func(uac string) {
			defer concurrent.Done()
			uacInfo, err := uacGenerator.GetUacInfo(uac)
			if err == datastore.ErrNoSuchEntity {
				uacGenerator.importMu.Lock()
				uacsToImport = append(uacsToImport, uac)
				uacGenerator.importMu.Unlock()
				return
			}
			if err != nil {
				uacGenerator.importMu.Lock()
				errors = append(errors, err)
				uacGenerator.importMu.Unlock()
				return
			}
			if uacInfo.InstrumentName == UNKNOWNINSTRUMENT {
				return
			}
			uacGenerator.importMu.Lock()
			importError.InstrumentUACs = append(importError.InstrumentUACs, uac)
			uacGenerator.importMu.Unlock()
		}(uac)
	}
	concurrent.WaitAllDone()

	if len(errors) > 0 {
		return nil, errors[0]
	}

	if importError.HasErrors() {
		return nil, &importError
	}
	return uacsToImport, nil
}

func (uacGenerator *UacGenerator) importUACs(uacs []string) (int, error) {
	var (
		updateCount = 0
		errors      []error
	)

	if len(uacs) == 0 {
		return 0, nil
	}

	concurrent := goccm.New(MAXCONCURRENT)
	for _, uac := range uacs {
		concurrent.Wait()
		go func(uac string) {
			defer concurrent.Done()
			err := uacGenerator.AddUacToDatastore(uac, UNKNOWNINSTRUMENT, UNKNOWNINSTRUMENT)
			if err != nil {
				uacGenerator.importMu.Lock()
				errors = append(errors, err)
				uacGenerator.importMu.Unlock()
				return
			}
			uacGenerator.importMu.Lock()
			updateCount++
			uacGenerator.importMu.Unlock()
		}(uac)
	}
	concurrent.WaitAllDone()

	if len(errors) > 0 {
		return 0, errors[0]
	}

	return updateCount, nil
}

func ChunkUAC(uac string) *UacChunks {
	var chunks []string
	runes := []rune(uac)

	if len(runes) == 0 {
		return nil
	}

	for i := 0; i < len(runes); i += 4 {
		nn := i + 4
		if nn > len(runes) {
			nn = len(runes)
		}
		chunks = append(chunks, string(runes[i:nn]))
	}
	uacChunks := &UacChunks{UAC1: chunks[0], UAC2: chunks[1], UAC3: chunks[2]}
	if len(chunks) >= 4 {
		uacChunks.UAC4 = chunks[3]
	}
	return uacChunks
}

func (uacGenerator *UacGenerator) adminDeleteChunk(uacKeyChunk []*datastore.Key, concurrent goccm.ConcurrencyManager) {
	defer concurrent.Done()
	err := uacGenerator.DatastoreClient.DeleteMulti(uacGenerator.Context, uacKeyChunk)
	if err != nil {
		log.Println(err)
	}
}

func (uacGenerator *UacGenerator) instrumentCaseQuery(instrumentName, caseID string) *datastore.Query {
	query := datastore.NewQuery(uacGenerator.UacKind)
	query = query.FilterField("instrument_name", "=", strings.ToLower(instrumentName))
	return query.FilterField(strings.ToLower("case_id"), "=", strings.ToLower(caseID))
}

func (uacGenerator *UacGenerator) instrumentQuery(instrumentName string) *datastore.Query {
	query := datastore.NewQuery(uacGenerator.UacKind)
	return query.FilterField("instrument_name", "=", strings.ToLower(instrumentName))
}

func (uacGenerator *UacGenerator) instrumentUacDisabledQuery(instrumentName string) *datastore.Query {
	query := datastore.NewQuery(uacGenerator.UacKind)

	query = query.FilterField("instrument_name", "=", strings.ToLower(instrumentName))
	return query.FilterField(strings.ToLower("disabled"), "=", true)
}

func (uacGenerator *UacGenerator) instrumentNamesQuery() *datastore.Query {
	query := datastore.NewQuery(uacGenerator.UacKind)
	query = query.Project("instrument_name")
	return query.DistinctOn("instrument_name")
}

func chunkDatastoreKeys(keys []*datastore.Key) [][]*datastore.Key {
	var (
		chunks    [][]*datastore.Key
		chunkSize = MAXCONCURRENT
	)
	for i := 0; i < len(keys); i += chunkSize {
		end := i + chunkSize

		if end > len(keys) {
			end = len(keys)
		}

		chunks = append(chunks, keys[i:end])
	}

	return chunks
}

func alreadyExistsError(err error) bool {
	if statusErr, ok := status.FromError(err); ok {
		return statusErr.Code() == codes.AlreadyExists
	}
	return false
}
