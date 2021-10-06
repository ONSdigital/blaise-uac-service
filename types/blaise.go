package types

import "net/http"

type BlaiseRestApi struct {
	BaseUrl    string
	Serverpark string
	Client     *http.Client
}

type InstrumentModes []string
