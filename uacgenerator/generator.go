package uacgenerator

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"sync"
	"time"

	"cloud.google.com/go/datastore"
	"github.com/zenthangplus/goccm"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	UACKIND       = "uac"
	MAXCONCURRENT = 500
)

//Generate mocks by running "go generate ./..."
//go:generate mockery --name Datastore
type Datastore interface {
	Mutate(context.Context, ...*datastore.Mutation) ([]*datastore.Key, error)
	GetAll(context.Context, *datastore.Query, interface{}) ([]*datastore.Key, error)
	Count(context.Context, *datastore.Query) (int, error)
	Get(context.Context, *datastore.Key, interface{}) error
	DeleteMulti(context.Context, []*datastore.Key) error
	Close() error
}

//Generate mocks by running "go generate ./..."
//go:generate mockery --name UacGeneratorInterface
type UacGeneratorInterface interface {
	Generate(string, []string) error
	GetAllUacs(string) (Uacs, error)
	GetAllUacsByCaseID(string) (Uacs, error)
	GetUacCount(string) (int, error)
	GetUacInfo(string) (*UacInfo, error)
	GetInstruments() ([]string, error)
	IncrementPostcodeAttempts(string) (*UacInfo, error)
	ResetPostcodeAttempts(string) (*UacInfo, error)
	AdminDelete(string) error
}

type UacGenerator struct {
	DatastoreClient Datastore
	Context         context.Context
	GenerateError   map[string]error
	mu              sync.Mutex
}

type UacChunks struct {
	UAC1 string `json:"uac1"`
	UAC2 string `json:"uac2"`
	UAC3 string `json:"uac3"`
}

type UacInfo struct {
	InstrumentName           string         `json:"instrument_name" datastore:"instrument_name"`
	CaseID                   string         `json:"case_id" datastore:"case_id"`
	PostcodeAttempts         int            `json:"postcode_attempts" datastore:"postcode_attempts"`
	PostcodeAttemptTimestamp string         `json:"postcode_attempt_timestamp" datastore:"postcode_attempt_timestamp"`
	UacChunks                *UacChunks     `json:"uac_chunks,omitempty" datastore:"-"`
	UAC                      *datastore.Key `json:"-" datastore:"__key__"`
	FullUAC                  string         `json:"-" datastore:"-"`
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

func (uacGenerator *UacGenerator) NewUac(instrumentName, caseID string, attempt int) (string, error) {
	if attempt >= 10 {
		return "", fmt.Errorf("Could not generate a unique UAC in 10 attempts")
	}
	uac := fmt.Sprintf("%012d", rand.Int63n(1e12))
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
	return datastore.NameKey(UACKIND, key, nil)
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

func (uacGenerator *UacGenerator) IncrementPostcodeAttempts(uac string) (*UacInfo, error) {
	uacInfo, err := uacGenerator.GetUacInfo(uac)
	if err != nil {
		return nil, err
	}
	uacInfo.PostcodeAttempts++
	uacInfo.PostcodeAttemptTimestamp = time.Now().UTC().String()

	newUACMutation := datastore.NewUpdate(uacGenerator.UacKey(uac), uacInfo)
	_, err = uacGenerator.DatastoreClient.Mutate(uacGenerator.Context, newUACMutation)
	return uacInfo, err
}

func (uacGenerator *UacGenerator) ResetPostcodeAttempts(uac string) (*UacInfo, error) {
	uacInfo, err := uacGenerator.GetUacInfo(uac)
	if err != nil {
		return nil, err
	}
	uacInfo.PostcodeAttempts = 0
	uacInfo.PostcodeAttemptTimestamp = ""

	newUACMutation := datastore.NewUpdate(uacGenerator.UacKey(uac), uacInfo)
	_, err = uacGenerator.DatastoreClient.Mutate(uacGenerator.Context, newUACMutation)
	return uacInfo, err
}

func (uacGenerator *UacGenerator) AdminDelete(instrumentName string) error {
	var instrumentUACs []*UacInfo
	instrumentUACKeys, err := uacGenerator.DatastoreClient.GetAll(uacGenerator.Context, uacGenerator.instrumentQuery(instrumentName), &instrumentUACs)
	if err != nil {
		return err
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
	return &UacChunks{UAC1: chunks[0], UAC2: chunks[1], UAC3: chunks[2]}
}

func (uacGenerator *UacGenerator) adminDeleteChunk(uacKeyChunk []*datastore.Key, concurrent goccm.ConcurrencyManager) {
	defer concurrent.Done()
	err := uacGenerator.DatastoreClient.DeleteMulti(uacGenerator.Context, uacKeyChunk)
	if err != nil {
		log.Println(err)
	}
}

func (uacGenerator *UacGenerator) instrumentCaseQuery(instrumentName, caseID string) *datastore.Query {
	query := datastore.NewQuery(UACKIND)
	query = query.Filter("instrument_name =", strings.ToLower(instrumentName))
	return query.Filter("case_id = ", strings.ToLower(caseID))
}

func (uacGenerator *UacGenerator) instrumentQuery(instrumentName string) *datastore.Query {
	query := datastore.NewQuery(UACKIND)
	return query.Filter("instrument_name =", strings.ToLower(instrumentName))
}

func (uacGenerator *UacGenerator) instrumentNamesQuery() *datastore.Query {
	query := datastore.NewQuery(UACKIND)
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
