// Code generated by mockery v2.53.3. DO NOT EDIT.

package mocks

import (
	models "github.com/spring-financial-group/peacock/pkg/models"
	mock "github.com/stretchr/testify/mock"
)

// FeathersUseCase is an autogenerated mock type for the FeathersUseCase type
type FeathersUseCase struct {
	mock.Mock
}

// GetFeathersFromBytes provides a mock function with given fields: data
func (_m *FeathersUseCase) GetFeathersFromBytes(data []byte) (*models.Feathers, error) {
	ret := _m.Called(data)

	if len(ret) == 0 {
		panic("no return value specified for GetFeathersFromBytes")
	}

	var r0 *models.Feathers
	var r1 error
	if rf, ok := ret.Get(0).(func([]byte) (*models.Feathers, error)); ok {
		return rf(data)
	}
	if rf, ok := ret.Get(0).(func([]byte) *models.Feathers); ok {
		r0 = rf(data)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*models.Feathers)
		}
	}

	if rf, ok := ret.Get(1).(func([]byte) error); ok {
		r1 = rf(data)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetFeathersFromFile provides a mock function with no fields
func (_m *FeathersUseCase) GetFeathersFromFile() (*models.Feathers, error) {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for GetFeathersFromFile")
	}

	var r0 *models.Feathers
	var r1 error
	if rf, ok := ret.Get(0).(func() (*models.Feathers, error)); ok {
		return rf()
	}
	if rf, ok := ret.Get(0).(func() *models.Feathers); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*models.Feathers)
		}
	}

	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ValidateFeathers provides a mock function with given fields: f
func (_m *FeathersUseCase) ValidateFeathers(f *models.Feathers) error {
	ret := _m.Called(f)

	if len(ret) == 0 {
		panic("no return value specified for ValidateFeathers")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(*models.Feathers) error); ok {
		r0 = rf(f)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// NewFeathersUseCase creates a new instance of FeathersUseCase. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewFeathersUseCase(t interface {
	mock.TestingT
	Cleanup(func())
}) *FeathersUseCase {
	mock := &FeathersUseCase{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
