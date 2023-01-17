// Code generated by mockery v2.15.0. DO NOT EDIT.

package mocks

import (
	context "context"

	github "github.com/google/go-github/v48/github"

	mock "github.com/stretchr/testify/mock"
)

// SCM is an autogenerated mock type for the SCM type
type SCM struct {
	mock.Mock
}

// CommentError provides a mock function with given fields: ctx, err
func (_m *SCM) CommentError(ctx context.Context, err error) error {
	ret := _m.Called(ctx, err)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, error) error); ok {
		r0 = rf(ctx, err)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// CommentOnPR provides a mock function with given fields: ctx, body
func (_m *SCM) CommentOnPR(ctx context.Context, body string) error {
	ret := _m.Called(ctx, body)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string) error); ok {
		r0 = rf(ctx, body)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// CreatePeacockCommitStatus provides a mock function with given fields: ctx, ref, state, statusContext
func (_m *SCM) CreatePeacockCommitStatus(ctx context.Context, ref string, state string, statusContext string) error {
	ret := _m.Called(ctx, ref, state, statusContext)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string, string) error); ok {
		r0 = rf(ctx, ref, state, statusContext)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// DeleteUsersComments provides a mock function with given fields: ctx
func (_m *SCM) DeleteUsersComments(ctx context.Context) error {
	ret := _m.Called(ctx)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context) error); ok {
		r0 = rf(ctx)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// GetFileFromBranch provides a mock function with given fields: ctx, branch, path
func (_m *SCM) GetFileFromBranch(ctx context.Context, branch string, path string) ([]byte, error) {
	ret := _m.Called(ctx, branch, path)

	var r0 []byte
	if rf, ok := ret.Get(0).(func(context.Context, string, string) []byte); ok {
		r0 = rf(ctx, branch, path)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]byte)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, string) error); ok {
		r1 = rf(ctx, branch, path)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetKey provides a mock function with given fields:
func (_m *SCM) GetKey() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// GetLatestCommitInBranch provides a mock function with given fields: ctx, branch
func (_m *SCM) GetLatestCommitInBranch(ctx context.Context, branch string) (*github.RepositoryCommit, error) {
	ret := _m.Called(ctx, branch)

	var r0 *github.RepositoryCommit
	if rf, ok := ret.Get(0).(func(context.Context, string) *github.RepositoryCommit); ok {
		r0 = rf(ctx, branch)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*github.RepositoryCommit)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, branch)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetPRComments provides a mock function with given fields: ctx
func (_m *SCM) GetPRComments(ctx context.Context) ([]*github.IssueComment, error) {
	ret := _m.Called(ctx)

	var r0 []*github.IssueComment
	if rf, ok := ret.Get(0).(func(context.Context) []*github.IssueComment); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*github.IssueComment)
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

// GetPRCommentsByUser provides a mock function with given fields: ctx
func (_m *SCM) GetPRCommentsByUser(ctx context.Context) ([]*github.IssueComment, error) {
	ret := _m.Called(ctx)

	var r0 []*github.IssueComment
	if rf, ok := ret.Get(0).(func(context.Context) []*github.IssueComment); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*github.IssueComment)
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

// GetPullRequestBodyFromCommit provides a mock function with given fields: ctx, sha
func (_m *SCM) GetPullRequestBodyFromCommit(ctx context.Context, sha string) (*string, error) {
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

// GetPullRequestBodyFromPRNumber provides a mock function with given fields: ctx
func (_m *SCM) GetPullRequestBodyFromPRNumber(ctx context.Context) (*string, error) {
	ret := _m.Called(ctx)

	var r0 *string
	if rf, ok := ret.Get(0).(func(context.Context) *string); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*string)
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

// HandleError provides a mock function with given fields: ctx, statusContext, headSHA, err
func (_m *SCM) HandleError(ctx context.Context, statusContext string, headSHA string, err error) error {
	ret := _m.Called(ctx, statusContext, headSHA, err)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string, error) error); ok {
		r0 = rf(ctx, statusContext, headSHA, err)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

type mockConstructorTestingTNewSCM interface {
	mock.TestingT
	Cleanup(func())
}

// NewSCM creates a new instance of SCM. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewSCM(t mockConstructorTestingTNewSCM) *SCM {
	mock := &SCM{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
