package blaiserestapi

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

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
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var caseIDs []string
	err = json.Unmarshal(body, &caseIDs)
	return caseIDs, err
}

func (blaiseRestApi *BlaiseRestApi) caseIdsUrl(instrumentName string) string {
	return fmt.Sprintf(
		"%s/api/v1/serverparks/%s/instruments/%s/cases/ids",
		blaiseRestApi.BaseUrl,
		blaiseRestApi.Serverpark,
		instrumentName,
	)
}
