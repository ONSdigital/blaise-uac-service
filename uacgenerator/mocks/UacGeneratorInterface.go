// Code generated by mockery v0.0.0-dev. DO NOT EDIT.

package mocks

import mock "github.com/stretchr/testify/mock"

// UacGeneratorInterface is an autogenerated mock type for the UacGeneratorInterface type
type UacGeneratorInterface struct {
	mock.Mock
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