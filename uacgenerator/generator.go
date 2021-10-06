package uacgenerator

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"sync"

	"github.com/ONSDigital/blaise-uac-service/types"

	"cloud.google.com/go/datastore"
	"github.com/zenthangplus/goccm"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	MAXCONCURRENT      = 500
	APPROVEDCHARACTERS = "bcdfghjklmnpqrstvxz23456789"
)

//Generate mocks by running "go generate ./..."
//go:generate mockery --name UacGeneratorInterface
type UacGeneratorInterface interface {
	Generate(string, []string) error
	GetAllUacs(string) (Uacs, error)
	GetAllUacsByCaseID(string) (Uacs, error)
	GetUacCount(string) (int, error)
	GetUacInfo(string) (*UacInfo, error)
	GetInstruments() ([]string, error)
	AdminDelete(string) error
}

type UacGenerator struct {
	UacKind         string
	DatastoreClient types.Datastore
	Context         context.Context
	GenerateError   map[string]error
	Randomizer      *rand.Rand
	mu              sync.Mutex
}

type UacInfo struct {
	InstrumentName string           `json:"instrument_name" datastore:"instrument_name"`
	CaseID         string           `json:"case_id" datastore:"case_id"`
	UacChunks      *types.UacChunks `json:"uac_chunks,omitempty" datastore:"-"`
	UAC            *datastore.Key   `json:"-" datastore:"__key__"`
	FullUAC        string           `json:"-" datastore:"-"`
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

func NewUacGenerator(datastoreClient types.Datastore, uacKind string) *UacGenerator {
	return &UacGenerator{
		UacKind:         uacKind,
		Context:         context.Background(),
		Randomizer:      rand.New(cryptoSource{}),
		DatastoreClient: datastoreClient,
	}
}

func (uacGenerator *UacGenerator) GenerateUac12() string {
	return fmt.Sprintf("%012d", uacGenerator.Randomizer.Int63n(1e12))
}

func (uacGenerator *UacGenerator) GenerateUac16() string {
	b := make([]byte, 16)
	for i := range b {
		b[i] = APPROVEDCHARACTERS[uacGenerator.Randomizer.Intn(len(APPROVEDCHARACTERS))]
	}
	return string(b)
}

func (uacGenerator *UacGenerator) NewUac(instrumentName, caseID string, attempt int) (string, error) {
	var uac string
	switch uacGenerator.UacKind {
	case "uac12":
		uac = uacGenerator.GenerateUac12()
	case "uac16":
		uac = uacGenerator.GenerateUac16()
	default:
		uac = ""
	}

	if uac == "" {
		return "", fmt.Errorf("Cannot generate UACs for invalid UacKind")
	}
	if caseID == "" {
		return "", fmt.Errorf("Cannot generate UACs for blank caseIDs")
	}
	if attempt >= 10 {
		return "", fmt.Errorf("Could not generate a unique UAC in 10 attempts")
	}

	uac, err := uacGenerator.DatastoreFunk(uac, instrumentName, caseID, attempt)
	if err != nil {
		return "", err
	}
	return uac, nil
}

func (uacGenerator *UacGenerator) DatastoreFunk(uac string, instrumentName, caseID string, attempt int) (string, error) {
	// Cannot workout how the hell to mock/ test this :(
	newUACMutation := datastore.NewInsert(uacGenerator.UacKey(uac), &UacInfo{
		InstrumentName: strings.ToLower(instrumentName),
		CaseID:         strings.ToLower(caseID),
	})
	_, err := uacGenerator.DatastoreClient.Mutate(uacGenerator.Context, newUACMutation)
	if err != nil {
		if statusErr, ok := status.FromError(err); ok {
			if statusErr.Code() == codes.AlreadyExists {
				return uacGenerator.NewUac(instrumentName, caseID, attempt+1)
			}
		}
		return "", err
	}
	return uac, nil
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
			uacGenerator.GenerateUniqueUac(instrumentName, caseID)
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

func (uacGenerator *UacGenerator) AdminDelete(instrumentName string) error {
	var instrumentUACs []*UacInfo
	instrumentUACKeys, err := uacGenerator.DatastoreClient.GetAll(uacGenerator.Context, uacGenerator.instrumentQuery(instrumentName), &instrumentUACs)
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

func ChunkUAC(uac string) *types.UacChunks {
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
	uacChunks := &types.UacChunks{UAC1: chunks[0], UAC2: chunks[1], UAC3: chunks[2]}
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
	query = query.Filter("instrument_name =", strings.ToLower(instrumentName))
	return query.Filter("case_id = ", strings.ToLower(caseID))
}

func (uacGenerator *UacGenerator) instrumentQuery(instrumentName string) *datastore.Query {
	query := datastore.NewQuery(uacGenerator.UacKind)
	return query.Filter("instrument_name =", strings.ToLower(instrumentName))
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
