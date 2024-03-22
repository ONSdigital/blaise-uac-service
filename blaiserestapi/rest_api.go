package blaiserestapi

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"log"
)

const CAWIMODE = "CAWI"

//Generate mocks by running "go generate ./..."
//go:generate mockery --name BlaiseRestApiInterface
type BlaiseRestApiInterface interface {
	GetCaseIds(string) ([]string, error)
	GetInstrumentModes(string) (InstrumentModes, error)
}

type InstrumentModes []string

type BlaiseRestApi struct {
	BaseUrl    string
	Serverpark string
	Client     *http.Client
}

func (blaiseRestApi *BlaiseRestApi) GetCaseIds(instrumentName string) ([]string, error) {
    log.Printf("UAC DEBUG: Calling blaiseRestApi.caseIdsUrl(%v)...", instrumentName)
	req, err := http.NewRequest("GET", blaiseRestApi.caseIdsUrl(instrumentName), nil)
	if err != nil {
        log.Printf("UAC DEBUG: blaiseRestApi.caseIdsUrl(%v) failed with the following error: %v", instrumentName, err)
		return nil, err
	}
    log.Println("UAC DEBUG: Calling blaiseRestApi.Client.Do(req)...")
	req.Header.Add("Accept", "application/json")
	resp, err := blaiseRestApi.Client.Do(req)
	if err != nil {
        log.Printf("UAC DEBUG: blaiseRestApi.Client.Do(req) failed with the following error: %v", err)
		return nil, err
	}

	defer resp.Body.Close()
    log.Println("UAC DEBUG: Validating resp.StatusCode...")
	if resp.StatusCode == http.StatusNotFound {
        log.Printf("UAC DEBUG: Validating resp.StatusCode failed with the following error: %v", err)
		return nil, fmt.Errorf("Instrument not found")
	}

    log.Println("UAC DEBUG: Getting body from io.ReadAll()...")
	body, err := io.ReadAll(resp.Body)
	if err != nil {
        log.Printf("UAC DEBUG: Getting body from io.ReadAll() failed with the following error: %v", err)
		return nil, err
	}

    log.Println("UAC DEBUG: Unmarshalling caseIDs to json...")
	var caseIDs []string
	err = json.Unmarshal(body, &caseIDs)
    log.Printf("UAC DEBUG: Returning caseIDs (%v) and err (%v)", caseIDs, err)

	return caseIDs, err
}

func (blaiseRestApi *BlaiseRestApi) GetInstrumentModes(instrumentName string) (InstrumentModes, error) {
	req, err := http.NewRequest("GET", blaiseRestApi.instrumentModeUrl(instrumentName), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Accept", "application/json")
	resp, err := blaiseRestApi.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("Instrument not found")
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var instrument_modes InstrumentModes
	err = json.Unmarshal(body, &instrument_modes)
	return instrument_modes, err
}

func (blaiseRestApi *BlaiseRestApi) caseIdsUrl(instrumentName string) string {
	return fmt.Sprintf(
		"%s/api/v2/serverparks/%s/questionnaires/%s/cases/ids",
		blaiseRestApi.BaseUrl,
		blaiseRestApi.Serverpark,
		instrumentName,
	)
}

func (blaiseRestApi *BlaiseRestApi) instrumentModeUrl(instrumentName string) string {
	return fmt.Sprintf(
		"%s/api/v2/serverparks/%s/questionnaires/%s/modes",
		blaiseRestApi.BaseUrl,
		blaiseRestApi.Serverpark,
		instrumentName,
	)
}

func (instrumentModes InstrumentModes) HasCawi() bool {
	for _, mode := range instrumentModes {
		if mode == CAWIMODE {
			return true
		}
	}
	return false
}
