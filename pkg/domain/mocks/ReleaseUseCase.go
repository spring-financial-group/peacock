// Code generated by mockery v2.42.0. DO NOT EDIT.

package mocks

import (
	context "context"

	mock "github.com/stretchr/testify/mock"

	models "github.com/spring-financial-group/peacock/pkg/models"

	time "time"
)

// ReleaseUseCase is an autogenerated mock type for the ReleaseUseCase type
type ReleaseUseCase struct {
	mock.Mock
}

// GetReleases provides a mock function with given fields: ctx, environment, startTime, teams
func (_m *ReleaseUseCase) GetReleases(ctx context.Context, environment string, startTime time.Time, teams []string) ([]models.Release, error) {
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

// SaveRelease provides a mock function with given fields: ctx, environment, releaseNotes, prSummary
func (_m *ReleaseUseCase) SaveRelease(ctx context.Context, environment string, releaseNotes []models.ReleaseNote, prSummary models.PullRequestSummary) error {
	ret := _m.Called(ctx, environment, releaseNotes, prSummary)

	if len(ret) == 0 {
		panic("no return value specified for SaveRelease")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, []models.ReleaseNote, models.PullRequestSummary) error); ok {
		r0 = rf(ctx, environment, releaseNotes, prSummary)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// NewReleaseUseCase creates a new instance of ReleaseUseCase. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewReleaseUseCase(t interface {
	mock.TestingT
	Cleanup(func())
}) *ReleaseUseCase {
	mock := &ReleaseUseCase{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
