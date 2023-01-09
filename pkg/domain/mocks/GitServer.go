// Code generated by mockery v2.15.0. DO NOT EDIT.

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

// CommentError provides a mock function with given fields: ctx, owner, repo, prNumber, err
func (_m *GitServer) CommentError(ctx context.Context, owner string, repo string, prNumber int, err error) error {
	ret := _m.Called(ctx, owner, repo, prNumber, err)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string, int, error) error); ok {
		r0 = rf(ctx, owner, repo, prNumber, err)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// CommentOnPR provides a mock function with given fields: ctx, owner, repo, prNumber, body
func (_m *GitServer) CommentOnPR(ctx context.Context, owner string, repo string, prNumber int, body string) error {
	ret := _m.Called(ctx, owner, repo, prNumber, body)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string, int, string) error); ok {
		r0 = rf(ctx, owner, repo, prNumber, body)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// DeleteUsersComments provides a mock function with given fields: ctx, owner, repo, user, prNumber
func (_m *GitServer) DeleteUsersComments(ctx context.Context, owner string, repo string, user string, prNumber int) error {
	ret := _m.Called(ctx, owner, repo, user, prNumber)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string, string, int) error); ok {
		r0 = rf(ctx, owner, repo, user, prNumber)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// GetFileFromBranch provides a mock function with given fields: ctx, owner, repo, branch, path
func (_m *GitServer) GetFileFromBranch(ctx context.Context, owner string, repo string, branch string, path string) ([]byte, error) {
	ret := _m.Called(ctx, owner, repo, branch, path)

	var r0 []byte
	if rf, ok := ret.Get(0).(func(context.Context, string, string, string, string) []byte); ok {
		r0 = rf(ctx, owner, repo, branch, path)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]byte)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, string, string, string) error); ok {
		r1 = rf(ctx, owner, repo, branch, path)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetPRComments provides a mock function with given fields: ctx, owner, repo, prNumber
func (_m *GitServer) GetPRComments(ctx context.Context, owner string, repo string, prNumber int) ([]*github.IssueComment, error) {
	ret := _m.Called(ctx, owner, repo, prNumber)

	var r0 []*github.IssueComment
	if rf, ok := ret.Get(0).(func(context.Context, string, string, int) []*github.IssueComment); ok {
		r0 = rf(ctx, owner, repo, prNumber)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*github.IssueComment)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, string, int) error); ok {
		r1 = rf(ctx, owner, repo, prNumber)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetPRCommentsByUser provides a mock function with given fields: ctx, owner, repo, user, prNumber
func (_m *GitServer) GetPRCommentsByUser(ctx context.Context, owner string, repo string, user string, prNumber int) ([]*github.IssueComment, error) {
	ret := _m.Called(ctx, owner, repo, user, prNumber)

	var r0 []*github.IssueComment
	if rf, ok := ret.Get(0).(func(context.Context, string, string, string, int) []*github.IssueComment); ok {
		r0 = rf(ctx, owner, repo, user, prNumber)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*github.IssueComment)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, string, string, int) error); ok {
		r1 = rf(ctx, owner, repo, user, prNumber)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetPullRequestBodyFromCommit provides a mock function with given fields: ctx, owner, repo, sha
func (_m *GitServer) GetPullRequestBodyFromCommit(ctx context.Context, owner string, repo string, sha string) (*string, error) {
	ret := _m.Called(ctx, owner, repo, sha)

	var r0 *string
	if rf, ok := ret.Get(0).(func(context.Context, string, string, string) *string); ok {
		r0 = rf(ctx, owner, repo, sha)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*string)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, string, string) error); ok {
		r1 = rf(ctx, owner, repo, sha)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetPullRequestBodyFromPRNumber provides a mock function with given fields: ctx, owner, repo, prNumber
func (_m *GitServer) GetPullRequestBodyFromPRNumber(ctx context.Context, owner string, repo string, prNumber int) (*string, error) {
	ret := _m.Called(ctx, owner, repo, prNumber)

	var r0 *string
	if rf, ok := ret.Get(0).(func(context.Context, string, string, int) *string); ok {
		r0 = rf(ctx, owner, repo, prNumber)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*string)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, string, int) error); ok {
		r1 = rf(ctx, owner, repo, prNumber)
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
