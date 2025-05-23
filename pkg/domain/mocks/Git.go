// Code generated by mockery v2.53.3. DO NOT EDIT.

package mocks

import mock "github.com/stretchr/testify/mock"

// Git is an autogenerated mock type for the Git type
type Git struct {
	mock.Mock
}

// GetLatestCommitSHA provides a mock function with given fields: dir
func (_m *Git) GetLatestCommitSHA(dir string) (string, error) {
	ret := _m.Called(dir)

	if len(ret) == 0 {
		panic("no return value specified for GetLatestCommitSHA")
	}

	var r0 string
	var r1 error
	if rf, ok := ret.Get(0).(func(string) (string, error)); ok {
		return rf(dir)
	}
	if rf, ok := ret.Get(0).(func(string) string); ok {
		r0 = rf(dir)
	} else {
		r0 = ret.Get(0).(string)
	}

	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(dir)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetRepoOwnerAndName provides a mock function with given fields: dir
func (_m *Git) GetRepoOwnerAndName(dir string) (string, string, error) {
	ret := _m.Called(dir)

	if len(ret) == 0 {
		panic("no return value specified for GetRepoOwnerAndName")
	}

	var r0 string
	var r1 string
	var r2 error
	if rf, ok := ret.Get(0).(func(string) (string, string, error)); ok {
		return rf(dir)
	}
	if rf, ok := ret.Get(0).(func(string) string); ok {
		r0 = rf(dir)
	} else {
		r0 = ret.Get(0).(string)
	}

	if rf, ok := ret.Get(1).(func(string) string); ok {
		r1 = rf(dir)
	} else {
		r1 = ret.Get(1).(string)
	}

	if rf, ok := ret.Get(2).(func(string) error); ok {
		r2 = rf(dir)
	} else {
		r2 = ret.Error(2)
	}

	return r0, r1, r2
}

// NewGit creates a new instance of Git. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewGit(t interface {
	mock.TestingT
	Cleanup(func())
}) *Git {
	mock := &Git{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
