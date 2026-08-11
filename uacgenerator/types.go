package uacgenerator

import "cloud.google.com/go/datastore"

const (
	MaxConcurrent      = 500
	maxUACAttempts     = 10
	ApprovedCharacters = "bcdfghjklmnpqrstvxz23456789"
	UnknownInstrument  = "unknown"
	UnknownCaseID      = "unknown"
)

type UACChunks struct {
	UAC1 string `json:"uac1"`
	UAC2 string `json:"uac2"`
	UAC3 string `json:"uac3"`
	UAC4 string `json:"uac4,omitempty"`
}

type UACInfo struct {
	InstrumentName string         `json:"instrument_name" datastore:"instrument_name"`
	CaseID         string         `json:"case_id" datastore:"case_id"`
	UACChunks      *UACChunks     `json:"uac_chunks,omitempty" datastore:"-"`
	UAC            *datastore.Key `json:"-" datastore:"__key__"`
	FullUAC        string         `json:"full_uac,omitempty" datastore:"-"`
	Disabled       bool           `json:"disabled" datastore:"disabled"`
}

type UACs map[string]*UACInfo

func (uacs UACs) BuildUACChunks() {
	for uac, uacInfo := range uacs {
		if uacInfo.FullUAC != "" {
			uac = uacInfo.FullUAC
		}
		uacInfo.UACChunks = chunkUAC(uac)
	}
}

func chunkUAC(uac string) *UACChunks {
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
	uacChunks := &UACChunks{UAC1: chunks[0], UAC2: chunks[1], UAC3: chunks[2]}
	if len(chunks) >= 4 {
		uacChunks.UAC4 = chunks[3]
	}
	return uacChunks
}
