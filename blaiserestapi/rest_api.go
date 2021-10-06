package blaiserestapi

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
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
	req, err := http.NewRequest("GET", blaiseRestApi.caseIdsUrl(instrumentName), nil)
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
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var caseIDs []string
	err = json.Unmarshal(body, &caseIDs)
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
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var instrument_modes InstrumentModes
	err = json.Unmarshal(body, &instrument_modes)
	return instrument_modes, err
}

func (blaiseRestApi *BlaiseRestApi) caseIdsUrl(instrumentName string) string {
	return fmt.Sprintf(
		"%s/api/v1/serverparks/%s/instruments/%s/cases/ids",
		blaiseRestApi.BaseUrl,
		blaiseRestApi.Serverpark,
		instrumentName,
	)
}

func (blaiseRestApi *BlaiseRestApi) instrumentModeUrl(instrumentName string) string {
	return fmt.Sprintf(
		"%s/api/v1/serverparks/%s/instruments/%s/modes",
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
