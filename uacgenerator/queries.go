package uacgenerator

import (
	"strings"

	"cloud.google.com/go/datastore"
)

func (uacService *UACService) instrumentCaseQuery(instrumentName, caseID string) *datastore.Query {
	query := datastore.NewQuery(uacService.UACKind)
	query = query.FilterField("instrument_name", "=", strings.ToLower(instrumentName))
	return query.FilterField("case_id", "=", strings.ToLower(caseID))
}

func (uacService *UACService) instrumentQuery(instrumentName string) *datastore.Query {
	query := datastore.NewQuery(uacService.UACKind)
	return query.FilterField("instrument_name", "=", strings.ToLower(instrumentName))
}

func (uacService *UACService) instrumentUACDisabledQuery(instrumentName string) *datastore.Query {
	query := datastore.NewQuery(uacService.UACKind)
	query = query.FilterField("instrument_name", "=", strings.ToLower(instrumentName))
	return query.FilterField("disabled", "=", true)
}

func (uacService *UACService) instrumentNamesQuery() *datastore.Query {
	query := datastore.NewQuery(uacService.UACKind)
	query = query.Project("instrument_name")
	return query.DistinctOn("instrument_name")
}

func chunkDatastoreKeys(keys []*datastore.Key) [][]*datastore.Key {
	var (
		chunks    [][]*datastore.Key
		chunkSize = MaxConcurrent
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
