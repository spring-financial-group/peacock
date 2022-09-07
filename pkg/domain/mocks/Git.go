// Code generated by mockery v2.14.0. DO NOT EDIT.

package mocks

import (
	context "context"

	github "github.com/google/go-github/v47/github"

	mock "github.com/stretchr/testify/mock"
)

// Git is an autogenerated mock type for the Git type
type Git struct {
	mock.Mock
}

// CommentOnPR provides a mock function with given fields: ctx, pullRequest, body
func (_m *Git) CommentOnPR(ctx context.Context, pullRequest *github.PullRequest, body string) error {
	ret := _m.Called(ctx, pullRequest, body)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *github.PullRequest, string) error); ok {
		r0 = rf(ctx, pullRequest, body)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// GetPullRequestFromLastCommit provides a mock function with given fields: ctx
func (_m *Git) GetPullRequestFromLastCommit(ctx context.Context) (*github.PullRequest, error) {
	ret := _m.Called(ctx)

	var r0 *github.PullRequest
	if rf, ok := ret.Get(0).(func(context.Context) *github.PullRequest); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*github.PullRequest)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(ctx)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetPullRequestFromPRNumber provides a mock function with given fields: ctx, prNumber
func (_m *Git) GetPullRequestFromPRNumber(ctx context.Context, prNumber int) (*github.PullRequest, error) {
	ret := _m.Called(ctx, prNumber)

	var r0 *github.PullRequest
	if rf, ok := ret.Get(0).(func(context.Context, int) *github.PullRequest); ok {
		r0 = rf(ctx, prNumber)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*github.PullRequest)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, int) error); ok {
		r1 = rf(ctx, prNumber)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

type mockConstructorTestingTNewGit interface {
	mock.TestingT
	Cleanup(func())
}

// NewGit creates a new instance of Git. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewGit(t mockConstructorTestingTNewGit) *Git {
	mock := &Git{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
