package uacgenerator

import (
	"context"
	"errors"
	"fmt"
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

var uac16Regex = regexp.MustCompile(`^[` + ApprovedCharacters + `]{16}$`)

type UACService struct {
	UACKind         string
	DatastoreClient DatastoreInterface
	randomizer      *rand.Rand
}

func NewUACService(datastoreClient DatastoreInterface, uacKind string) *UACService {
	return &UACService{
		UACKind:         uacKind,
		randomizer:      rand.New(cryptoSource{}),
		DatastoreClient: datastoreClient,
	}
}

func (uacService *UACService) generateUAC12() string {
	var uac string
	for i := 0; i < 3; i++ {
		uacSegment := uacService.randomizer.Int63n(9999 - 1000)
		uac = fmt.Sprintf("%s%d", uac, uacSegment+1000)
	}
	return uac
}

func (uacService *UACService) generateUAC16() string {
	b := make([]byte, 16)
	for i := range b {
		b[i] = ApprovedCharacters[uacService.randomizer.Intn(len(ApprovedCharacters))]
	}
	return string(b)
}

func (uacService *UACService) newUAC(ctx context.Context, instrumentName, caseID string) (string, error) {
	if caseID == "" {
		return "", ErrBlankCaseID
	}

	for currentAttempt := 0; currentAttempt < maxUACAttempts; currentAttempt++ {
		var uac string
		switch uacService.UACKind {
		case "uac":
			uac = uacService.generateUAC12()
		case "uac16":
			uac = uacService.generateUAC16()
		default:
			return "", ErrInvalidUACKind
		}

		err := uacService.addUACToDatastore(ctx, uac, instrumentName, caseID)
		if err == nil {
			return uac, nil
		}
		if !alreadyExistsError(err) {
			return "", err
		}
	}

	return "", ErrCouldNotGenerateUniqueUAC
}

func (uacService *UACService) addUACToDatastore(ctx context.Context, uac string, instrumentName, caseID string) error {
	newUACMutation := datastore.NewInsert(uacService.uacKey(uac), &UACInfo{
		InstrumentName: strings.ToLower(instrumentName),
		CaseID:         strings.ToLower(caseID),
	})
	_, err := uacService.DatastoreClient.Mutate(ctx, newUACMutation)
	return err
}

func (uacService *UACService) uacKey(key string) *datastore.Key {
	return datastore.NameKey(uacService.UACKind, key, nil)
}

func (uacService *UACService) uacExistsForCase(ctx context.Context, instrumentName, caseID string) (bool, error) {
	var existingUACs []*UACInfo
	existingUACKeys, err := uacService.DatastoreClient.GetAll(
		ctx,
		uacService.instrumentCaseQuery(instrumentName, caseID),
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

func (uacService *UACService) generateUniqueUAC(ctx context.Context, instrumentName, caseID string) error {
	exists, err := uacService.uacExistsForCase(ctx, instrumentName, caseID)
	if err != nil {
		return err
	}
	if !exists {
		_, err := uacService.newUAC(ctx, instrumentName, caseID)
		if err != nil {
			return err
		}
	}
	return nil
}

func (uacService *UACService) Generate(ctx context.Context, instrumentName string, caseIDs []string) error {
	if len(caseIDs) == 0 {
		return nil
	}
	errCh := make(chan error, len(caseIDs))
	concurrent := goccm.New(MaxConcurrent)
	for _, caseID := range caseIDs {
		concurrent.Wait()
		go func(caseID string) {
			defer concurrent.Done()
			err := uacService.generateUniqueUAC(ctx, instrumentName, caseID)
			if err != nil {
				errCh <- err
			}
		}(caseID)
	}
	concurrent.WaitAllDone()
	close(errCh)
	for err := range errCh {
		return err
	}
	return nil
}

func (uacService *UACService) GetAllUACs(ctx context.Context, instrumentName string) (UACs, error) {
	var uacInfos []*UACInfo
	_, err := uacService.DatastoreClient.GetAll(ctx, uacService.instrumentQuery(instrumentName), &uacInfos)
	if err != nil {
		return nil, err
	}
	uacs := make(UACs)
	for _, uacInfo := range uacInfos {
		uacs[uacInfo.UAC.Name] = uacInfo
	}
	return uacs, nil
}

func (uacService *UACService) GetAllUACsByCaseID(ctx context.Context, instrumentName string) (UACs, error) {
	var uacInfos []*UACInfo
	_, err := uacService.DatastoreClient.GetAll(ctx, uacService.instrumentQuery(instrumentName), &uacInfos)
	if err != nil {
		return nil, err
	}
	uacs := make(UACs)
	for _, uacInfo := range uacInfos {
		uacInfo.FullUAC = uacInfo.UAC.Name
		uacs[uacInfo.CaseID] = uacInfo
	}
	if len(uacs) != len(uacInfos) {
		return nil, fmt.Errorf("fewer case IDs than UACs: duplicate case IDs detected")
	}
	return uacs, nil
}

func (uacService *UACService) GetAllDisabledUACs(ctx context.Context, instrumentName string) (UACs, error) {
	var uacInfos []*UACInfo
	_, err := uacService.DatastoreClient.GetAll(ctx, uacService.instrumentUACDisabledQuery(instrumentName), &uacInfos)
	if err != nil {
		return nil, err
	}
	uacs := make(UACs)
	for _, uacInfo := range uacInfos {
		uacInfo.FullUAC = uacInfo.UAC.Name
		uacs[uacInfo.CaseID] = uacInfo
	}
	if len(uacs) != len(uacInfos) {
		return nil, fmt.Errorf("fewer case IDs than UACs: duplicate case IDs detected")
	}
	return uacs, nil
}

func (uacService *UACService) DisableUAC(ctx context.Context, uac string) error {
	uacInfo := &UACInfo{}
	err := uacService.DatastoreClient.Get(ctx, uacService.uacKey(uac), uacInfo)
	if err != nil {
		if errors.Is(err, datastore.ErrNoSuchEntity) {
			return ErrInvalidUAC
		}
		return err
	}
	newUACMutation := datastore.NewUpdate(uacService.uacKey(uac), &UACInfo{
		InstrumentName: strings.ToLower(uacInfo.InstrumentName),
		CaseID:         strings.ToLower(uacInfo.CaseID),
		Disabled:       true,
	})
	_, err = uacService.DatastoreClient.Mutate(ctx, newUACMutation)
	if err != nil {
		return err
	}
	return nil
}

func (uacService *UACService) EnableUAC(ctx context.Context, uac string) error {
	uacInfo := &UACInfo{}
	err := uacService.DatastoreClient.Get(ctx, uacService.uacKey(uac), uacInfo)
	if err != nil {
		if errors.Is(err, datastore.ErrNoSuchEntity) {
			return ErrInvalidUAC
		}
		return err
	}
	newUACMutation := datastore.NewUpdate(uacService.uacKey(uac), &UACInfo{
		InstrumentName: strings.ToLower(uacInfo.InstrumentName),
		CaseID:         strings.ToLower(uacInfo.CaseID),
		Disabled:       false,
	})
	_, err = uacService.DatastoreClient.Mutate(ctx, newUACMutation)
	if err != nil {
		return err
	}
	return nil
}

func (uacService *UACService) GetUACCount(ctx context.Context, instrumentName string) (int, error) {
	return uacService.DatastoreClient.Count(ctx, uacService.instrumentQuery(instrumentName))
}

func (uacService *UACService) GetUACInfo(ctx context.Context, uac string) (*UACInfo, error) {
	uacInfo := &UACInfo{}
	err := uacService.DatastoreClient.Get(ctx, uacService.uacKey(uac), uacInfo)
	if err != nil {
		if errors.Is(err, datastore.ErrNoSuchEntity) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return uacInfo, nil
}

func (uacService *UACService) GetInstruments(ctx context.Context) ([]string, error) {
	var (
		uacInfos        []*UACInfo
		instrumentNames []string
	)
	_, err := uacService.DatastoreClient.GetAll(ctx, uacService.instrumentNamesQuery(), &uacInfos)
	if err != nil {
		return nil, err
	}
	for _, uacInfo := range uacInfos {
		instrumentNames = append(instrumentNames, uacInfo.InstrumentName)
	}
	return instrumentNames, nil
}

func (uacService *UACService) ImportUACs(ctx context.Context, uacs []string) (int, error) {
	if err := uacService.validateUACs(uacs); err != nil {
		return 0, err
	}
	uacsToImport, err := uacService.getUACsToImport(ctx, uacs)
	if err != nil {
		return 0, err
	}
	return uacService.importUACs(ctx, uacsToImport)
}

func (uacService *UACService) validateUAC12(uac string) bool {
	if len(uac) != 12 {
		return false
	}
	chunkedUAC := chunkUAC(uac)
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

func (uacService *UACService) validateUAC16(uac string) bool {
	return uac16Regex.MatchString(uac)
}

func (uacService *UACService) validateUAC(uac string) bool {
	if uacService.UACKind == "uac16" {
		return uacService.validateUAC16(uac)
	}
	return uacService.validateUAC12(uac)
}

func (uacService *UACService) validateUACs(uacs []string) error {
	var importError ImportError
	for _, uac := range uacs {
		if !uacService.validateUAC(uac) {
			importError.InvalidUACs = append(importError.InvalidUACs, uac)
		}
	}
	if importError.hasErrors() {
		return &importError
	}
	return nil
}

func (uacService *UACService) AdminDelete(ctx context.Context, instrumentName string) error {
	instrumentUACKeys, err := uacService.DatastoreClient.GetAll(ctx, uacService.instrumentQuery(instrumentName).KeysOnly(), nil)
	if err != nil {
		return err
	}
	if len(instrumentUACKeys) == 0 {
		return nil
	}
	uacKeyChunks := chunkDatastoreKeys(instrumentUACKeys)
	errCh := make(chan error, len(uacKeyChunks))
	concurrent := goccm.New(MaxConcurrent)
	for _, uacKeyChunk := range uacKeyChunks {
		concurrent.Wait()
		go func(uacKeyChunk []*datastore.Key) {
			defer concurrent.Done()
			if err := uacService.DatastoreClient.DeleteMulti(ctx, uacKeyChunk); err != nil {
				errCh <- err
			}
		}(uacKeyChunk)
	}
	concurrent.WaitAllDone()
	close(errCh)
	for err := range errCh {
		return err
	}
	return nil
}

func (uacService *UACService) getUACsToImport(ctx context.Context, uacs []string) ([]string, error) {
	var (
		uacsToImport []string
		importError  ImportError
		importMu     sync.Mutex
	)

	if len(uacs) == 0 {
		return nil, nil
	}

	errCh := make(chan error, len(uacs))
	concurrent := goccm.New(MaxConcurrent)
	for _, uac := range uacs {
		concurrent.Wait()
		go func(uac string) {
			defer concurrent.Done()
			uacInfo, err := uacService.GetUACInfo(ctx, uac)
			if errors.Is(err, ErrNotFound) {
				importMu.Lock()
				uacsToImport = append(uacsToImport, uac)
				importMu.Unlock()
				return
			}
			if err != nil {
				errCh <- err
				return
			}
			if uacInfo.InstrumentName == UnknownInstrument {
				return
			}
			importMu.Lock()
			importError.InstrumentUACs = append(importError.InstrumentUACs, uac)
			importMu.Unlock()
		}(uac)
	}
	concurrent.WaitAllDone()
	close(errCh)
	for err := range errCh {
		return nil, err
	}

	if importError.hasErrors() {
		return nil, &importError
	}
	return uacsToImport, nil
}

func (uacService *UACService) importUACs(ctx context.Context, uacs []string) (int, error) {
	var (
		updateCount = 0
		importMu    sync.Mutex
	)

	if len(uacs) == 0 {
		return 0, nil
	}

	errCh := make(chan error, len(uacs))
	concurrent := goccm.New(MaxConcurrent)
	for _, uac := range uacs {
		concurrent.Wait()
		go func(uac string) {
			defer concurrent.Done()
			err := uacService.addUACToDatastore(ctx, uac, UnknownInstrument, UnknownCaseID)
			if err != nil {
				errCh <- err
				return
			}
			importMu.Lock()
			updateCount++
			importMu.Unlock()
		}(uac)
	}
	concurrent.WaitAllDone()
	close(errCh)
	for err := range errCh {
		return 0, err
	}

	return updateCount, nil
}

func alreadyExistsError(err error) bool {
	if statusErr, ok := status.FromError(err); ok {
		return statusErr.Code() == codes.AlreadyExists
	}
	return false
}
