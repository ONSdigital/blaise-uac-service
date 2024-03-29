// Code generated by mockery v0.0.0-dev. DO NOT EDIT.

package mocks

import (
	blaiserestapi "github.com/ONSDigital/blaise-uac-service/blaiserestapi"
	mock "github.com/stretchr/testify/mock"
)

// BlaiseRestApiInterface is an autogenerated mock type for the BlaiseRestApiInterface type
type BlaiseRestApiInterface struct {
	mock.Mock
}

// GetCaseIds provides a mock function with given fields: _a0
func (_m *BlaiseRestApiInterface) GetCaseIds(_a0 string) ([]string, error) {
	ret := _m.Called(_a0)

	var r0 []string
	if rf, ok := ret.Get(0).(func(string) []string); ok {
		r0 = rf(_a0)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]string)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(_a0)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetInstrumentModes provides a mock function with given fields: _a0
func (_m *BlaiseRestApiInterface) GetInstrumentModes(_a0 string) (blaiserestapi.InstrumentModes, error) {
	ret := _m.Called(_a0)

	var r0 blaiserestapi.InstrumentModes
	if rf, ok := ret.Get(0).(func(string) blaiserestapi.InstrumentModes); ok {
		r0 = rf(_a0)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(blaiserestapi.InstrumentModes)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(_a0)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
