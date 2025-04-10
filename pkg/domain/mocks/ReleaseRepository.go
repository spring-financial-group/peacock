// Code generated by mockery v2.53.3. DO NOT EDIT.

package mocks

import (
	context "context"

	mock "github.com/stretchr/testify/mock"

	models "github.com/spring-financial-group/peacock/pkg/models"

	time "time"
)

// ReleaseRepository is an autogenerated mock type for the ReleaseRepository type
type ReleaseRepository struct {
	mock.Mock
}

// GetReleases provides a mock function with given fields: ctx, environment, startTime, teams
func (_m *ReleaseRepository) GetReleases(ctx context.Context, environment string, startTime time.Time, teams []string) ([]models.Release, error) {
	ret := _m.Called(ctx, environment, startTime, teams)

	if len(ret) == 0 {
		panic("no return value specified for GetReleases")
	}

	var r0 []models.Release
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, time.Time, []string) ([]models.Release, error)); ok {
		return rf(ctx, environment, startTime, teams)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, time.Time, []string) []models.Release); ok {
		r0 = rf(ctx, environment, startTime, teams)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]models.Release)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, time.Time, []string) error); ok {
		r1 = rf(ctx, environment, startTime, teams)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Insert provides a mock function with given fields: ctx, release
func (_m *ReleaseRepository) Insert(ctx context.Context, release models.Release) error {
	ret := _m.Called(ctx, release)

	if len(ret) == 0 {
		panic("no return value specified for Insert")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, models.Release) error); ok {
		r0 = rf(ctx, release)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// NewReleaseRepository creates a new instance of ReleaseRepository. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewReleaseRepository(t interface {
	mock.TestingT
	Cleanup(func())
}) *ReleaseRepository {
	mock := &ReleaseRepository{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
