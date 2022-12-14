// Code generated by mockery v2.14.0. DO NOT EDIT.

package mocks

import (
	context "context"

	github "github.com/google/go-github/v47/github"

	mock "github.com/stretchr/testify/mock"
)

// GitServer is an autogenerated mock type for the GitServer type
type GitServer struct {
	mock.Mock
}

// CommentOnPR provides a mock function with given fields: ctx, prNumber, body
func (_m *GitServer) CommentOnPR(ctx context.Context, prNumber int, body string) error {
	ret := _m.Called(ctx, prNumber, body)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, int, string) error); ok {
		r0 = rf(ctx, prNumber, body)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// GetPRComments provides a mock function with given fields: ctx, prNumber
func (_m *GitServer) GetPRComments(ctx context.Context, prNumber int) ([]*github.IssueComment, error) {
	ret := _m.Called(ctx, prNumber)

	var r0 []*github.IssueComment
	if rf, ok := ret.Get(0).(func(context.Context, int) []*github.IssueComment); ok {
		r0 = rf(ctx, prNumber)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*github.IssueComment)
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

// GetPullRequestBodyFromCommit provides a mock function with given fields: ctx, sha
func (_m *GitServer) GetPullRequestBodyFromCommit(ctx context.Context, sha string) (*string, error) {
	ret := _m.Called(ctx, sha)

	var r0 *string
	if rf, ok := ret.Get(0).(func(context.Context, string) *string); ok {
		r0 = rf(ctx, sha)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*string)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, sha)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetPullRequestBodyFromPRNumber provides a mock function with given fields: ctx, prNumber
func (_m *GitServer) GetPullRequestBodyFromPRNumber(ctx context.Context, prNumber int) (*string, error) {
	ret := _m.Called(ctx, prNumber)

	var r0 *string
	if rf, ok := ret.Get(0).(func(context.Context, int) *string); ok {
		r0 = rf(ctx, prNumber)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*string)
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

type mockConstructorTestingTNewGitServer interface {
	mock.TestingT
	Cleanup(func())
}

// NewGitServer creates a new instance of GitServer. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewGitServer(t mockConstructorTestingTNewGitServer) *GitServer {
	mock := &GitServer{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
