package blaiserestapi

//go:generate mockery

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const CAWIMode = "CAWI"

type BlaiseRESTAPIInterface interface {
	GetCaseIDs(string) ([]string, error)
	GetInstrumentModes(string) (InstrumentModes, error)
}

type InstrumentModes []string

type BlaiseRESTAPI struct {
	BaseURL    string
	Serverpark string
	Client     *http.Client
}

func (blaiseRESTAPI *BlaiseRESTAPI) GetCaseIDs(instrumentName string) ([]string, error) {
	req, err := http.NewRequest(http.MethodGet, blaiseRESTAPI.caseIDsURL(instrumentName), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Accept", "application/json")
	resp, err := blaiseRESTAPI.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return nil, ErrInstrumentNotFound
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var caseIDs []string
	err = json.Unmarshal(body, &caseIDs)
	return caseIDs, err
}

func (blaiseRESTAPI *BlaiseRESTAPI) GetInstrumentModes(instrumentName string) (InstrumentModes, error) {
	req, err := http.NewRequest(http.MethodGet, blaiseRESTAPI.instrumentModeURL(instrumentName), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Accept", "application/json")
	resp, err := blaiseRESTAPI.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return nil, ErrInstrumentNotFound
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var instrumentModes InstrumentModes
	err = json.Unmarshal(body, &instrumentModes)
	return instrumentModes, err
}

func (blaiseRESTAPI *BlaiseRESTAPI) caseIDsURL(instrumentName string) string {
	return fmt.Sprintf(
		"%s/api/v2/serverparks/%s/questionnaires/%s/cases/ids",
		blaiseRESTAPI.BaseURL,
		blaiseRESTAPI.Serverpark,
		instrumentName,
	)
}

func (blaiseRESTAPI *BlaiseRESTAPI) instrumentModeURL(instrumentName string) string {
	return fmt.Sprintf(
		"%s/api/v2/serverparks/%s/questionnaires/%s/modes",
		blaiseRESTAPI.BaseURL,
		blaiseRESTAPI.Serverpark,
		instrumentName,
	)
}

func (instrumentModes InstrumentModes) HasCAWI() bool {
	for _, mode := range instrumentModes {
		if mode == CAWIMode {
			return true
		}
	}
	return false
}
