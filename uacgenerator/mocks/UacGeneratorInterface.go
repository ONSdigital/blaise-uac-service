// Code generated by mockery v0.0.0-dev. DO NOT EDIT.

package mocks

import (
	uacgenerator "github.com/ONSDigital/blaise-uac-service/uacgenerator"
	mock "github.com/stretchr/testify/mock"
)

// UacGeneratorInterface is an autogenerated mock type for the UacGeneratorInterface type
type UacGeneratorInterface struct {
	mock.Mock
}

// AdminDelete provides a mock function with given fields: _a0
func (_m *UacGeneratorInterface) AdminDelete(_a0 string) error {
	ret := _m.Called(_a0)

	var r0 error
	if rf, ok := ret.Get(0).(func(string) error); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Generate provides a mock function with given fields: _a0, _a1
func (_m *UacGeneratorInterface) Generate(_a0 string, _a1 []string) error {
	ret := _m.Called(_a0, _a1)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, []string) error); ok {
		r0 = rf(_a0, _a1)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// GetAllUacs provides a mock function with given fields: _a0
func (_m *UacGeneratorInterface) GetAllUacs(_a0 string) (uacgenerator.Uacs, error) {
	ret := _m.Called(_a0)

	var r0 uacgenerator.Uacs
	if rf, ok := ret.Get(0).(func(string) uacgenerator.Uacs); ok {
		r0 = rf(_a0)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(uacgenerator.Uacs)
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

// GetAllUacsByCaseID provides a mock function with given fields: _a0
func (_m *UacGeneratorInterface) GetAllUacsByCaseID(_a0 string) (uacgenerator.Uacs, error) {
	ret := _m.Called(_a0)

	var r0 uacgenerator.Uacs
	if rf, ok := ret.Get(0).(func(string) uacgenerator.Uacs); ok {
		r0 = rf(_a0)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(uacgenerator.Uacs)
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

// GetInstruments provides a mock function with given fields:
func (_m *UacGeneratorInterface) GetInstruments() ([]string, error) {
	ret := _m.Called()

	var r0 []string
	if rf, ok := ret.Get(0).(func() []string); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]string)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetUacCount provides a mock function with given fields: _a0
func (_m *UacGeneratorInterface) GetUacCount(_a0 string) (int, error) {
	ret := _m.Called(_a0)

	var r0 int
	if rf, ok := ret.Get(0).(func(string) int); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Get(0).(int)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(_a0)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetUacInfo provides a mock function with given fields: _a0
func (_m *UacGeneratorInterface) GetUacInfo(_a0 string) (*uacgenerator.UacInfo, error) {
	ret := _m.Called(_a0)

	var r0 *uacgenerator.UacInfo
	if rf, ok := ret.Get(0).(func(string) *uacgenerator.UacInfo); ok {
		r0 = rf(_a0)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*uacgenerator.UacInfo)
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

// ImportUACs provides a mock function with given fields: _a0
func (_m *UacGeneratorInterface) ImportUACs(_a0 []string) (int, error) {
	ret := _m.Called(_a0)

	var r0 int
	if rf, ok := ret.Get(0).(func([]string) int); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Get(0).(int)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func([]string) error); ok {
		r1 = rf(_a0)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
