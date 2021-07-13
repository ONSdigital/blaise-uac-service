// Code generated by mockery v0.0.0-dev. DO NOT EDIT.

package mocks

import (
	context "context"

	datastore "cloud.google.com/go/datastore"
	mock "github.com/stretchr/testify/mock"
)

// Datastore is an autogenerated mock type for the Datastore type
type Datastore struct {
	mock.Mock
}

// Close provides a mock function with given fields:
func (_m *Datastore) Close() error {
	ret := _m.Called()

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// GetAll provides a mock function with given fields: _a0, _a1, _a2
func (_m *Datastore) GetAll(_a0 context.Context, _a1 *datastore.Query, _a2 interface{}) ([]*datastore.Key, error) {
	ret := _m.Called(_a0, _a1, _a2)

	var r0 []*datastore.Key
	if rf, ok := ret.Get(0).(func(context.Context, *datastore.Query, interface{}) []*datastore.Key); ok {
		r0 = rf(_a0, _a1, _a2)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*datastore.Key)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *datastore.Query, interface{}) error); ok {
		r1 = rf(_a0, _a1, _a2)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Mutate provides a mock function with given fields: _a0, _a1
func (_m *Datastore) Mutate(_a0 context.Context, _a1 ...*datastore.Mutation) ([]*datastore.Key, error) {
	_va := make([]interface{}, len(_a1))
	for _i := range _a1 {
		_va[_i] = _a1[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, _a0)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 []*datastore.Key
	if rf, ok := ret.Get(0).(func(context.Context, ...*datastore.Mutation) []*datastore.Key); ok {
		r0 = rf(_a0, _a1...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*datastore.Key)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, ...*datastore.Mutation) error); ok {
		r1 = rf(_a0, _a1...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
