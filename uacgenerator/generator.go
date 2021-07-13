package uacgenerator

import (
	"context"
	"fmt"
	"math/rand"
	"strings"

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
	Close() error
}

//Generate mocks by running "go generate ./..."
//go:generate mockery --name UacGeneratorInterface
type UacGeneratorInterface interface {
	Generate(string, []string) error
	GetAllUacs(string) (map[string]*UacInfo, error)
}

type UacGenerator struct {
	DatastoreClient Datastore
	Context         context.Context
}

type UacInfo struct {
	InstrumentName string         `json:"instrument_name" datastore:"instrument_name"`
	CaseID         string         `json:"case_id" datastore:"case_id"`
	UAC            *datastore.Key `json:"-" datastore:"__key__"`
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

func (uacGenerator *UacGenerator) GenerateUniqueUac(instrumentName, caseID string, concurrent goccm.ConcurrencyManager) error {
	defer concurrent.Done()
	exists, err := uacGenerator.UacExistsForCase(instrumentName, caseID)
	if err != nil {
		return err
	}
	if !exists {
		_, err := uacGenerator.NewUac(instrumentName, caseID, 0)
		if err != nil {
			return err
		}
	}
	return nil
}

func (uacGenerator *UacGenerator) Generate(instrumentName string, caseIDs []string) error {
	concurrent := goccm.New(MAXCONCURRENT)
	for _, caseID := range caseIDs {
		concurrent.Wait()
		// How do I check for errors from the goroutine?
		go uacGenerator.GenerateUniqueUac(instrumentName, caseID, concurrent)
	}
	concurrent.WaitAllDone()
	return nil
}

func (uacGenerator *UacGenerator) GetAllUacs(instrumentName string) (map[string]*UacInfo, error) {
	var uacInfos []*UacInfo
	_, err := uacGenerator.DatastoreClient.GetAll(uacGenerator.Context, uacGenerator.instrumentQuery(instrumentName), &uacInfos)
	if err != nil {
		return nil, err
	}
	uacs := make(map[string]*UacInfo)
	for _, uacInfo := range uacInfos {
		uacs[uacInfo.UAC.Name] = uacInfo
	}
	return uacs, nil
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
