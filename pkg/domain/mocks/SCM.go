// Code generated by mockery v2.53.3. DO NOT EDIT.

package mocks

import (
	context "context"

	github "github.com/google/go-github/v48/github"
	domain "github.com/spring-financial-group/peacock/pkg/domain"

	mock "github.com/stretchr/testify/mock"
)

// SCM is an autogenerated mock type for the SCM type
type SCM struct {
	mock.Mock
}

// CommentError provides a mock function with given fields: ctx, owner, repoName, prNumber, prOwner, err
func (_m *SCM) CommentError(ctx context.Context, owner string, repoName string, prNumber int, prOwner string, err error) error {
	ret := _m.Called(ctx, owner, repoName, prNumber, prOwner, err)

	if len(ret) == 0 {
		panic("no return value specified for CommentError")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string, int, string, error) error); ok {
		r0 = rf(ctx, owner, repoName, prNumber, prOwner, err)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// CommentOnPR provides a mock function with given fields: ctx, owner, repoName, prNumber, body
func (_m *SCM) CommentOnPR(ctx context.Context, owner string, repoName string, prNumber int, body string) error {
	ret := _m.Called(ctx, owner, repoName, prNumber, body)

	if len(ret) == 0 {
		panic("no return value specified for CommentOnPR")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string, int, string) error); ok {
		r0 = rf(ctx, owner, repoName, prNumber, body)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// CreatePeacockCommitStatus provides a mock function with given fields: ctx, owner, repoName, ref, state, statusContext
func (_m *SCM) CreatePeacockCommitStatus(ctx context.Context, owner string, repoName string, ref string, state domain.State, statusContext string) error {
	ret := _m.Called(ctx, owner, repoName, ref, state, statusContext)

	if len(ret) == 0 {
		panic("no return value specified for CreatePeacockCommitStatus")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string, string, domain.State, string) error); ok {
		r0 = rf(ctx, owner, repoName, ref, state, statusContext)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// DeleteUsersComments provides a mock function with given fields: ctx, owner, repoName, prNumber
func (_m *SCM) DeleteUsersComments(ctx context.Context, owner string, repoName string, prNumber int) error {
	ret := _m.Called(ctx, owner, repoName, prNumber)

	if len(ret) == 0 {
		panic("no return value specified for DeleteUsersComments")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string, int) error); ok {
		r0 = rf(ctx, owner, repoName, prNumber)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// GetFileFromBranch provides a mock function with given fields: ctx, owner, repoName, branch, path
func (_m *SCM) GetFileFromBranch(ctx context.Context, owner string, repoName string, branch string, path string) ([]byte, error) {
	ret := _m.Called(ctx, owner, repoName, branch, path)

	if len(ret) == 0 {
		panic("no return value specified for GetFileFromBranch")
	}

	var r0 []byte
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string, string, string) ([]byte, error)); ok {
		return rf(ctx, owner, repoName, branch, path)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, string, string, string) []byte); ok {
		r0 = rf(ctx, owner, repoName, branch, path)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]byte)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, string, string, string) error); ok {
		r1 = rf(ctx, owner, repoName, branch, path)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetFilesChangedFromPR provides a mock function with given fields: ctx, owner, repoName, prNumber
func (_m *SCM) GetFilesChangedFromPR(ctx context.Context, owner string, repoName string, prNumber int) ([]*github.CommitFile, error) {
	ret := _m.Called(ctx, owner, repoName, prNumber)

	if len(ret) == 0 {
		panic("no return value specified for GetFilesChangedFromPR")
	}

	var r0 []*github.CommitFile
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string, int) ([]*github.CommitFile, error)); ok {
		return rf(ctx, owner, repoName, prNumber)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, string, int) []*github.CommitFile); ok {
		r0 = rf(ctx, owner, repoName, prNumber)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*github.CommitFile)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, string, int) error); ok {
		r1 = rf(ctx, owner, repoName, prNumber)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetLatestCommitSHAInBranch provides a mock function with given fields: ctx, owner, repoName, branch
func (_m *SCM) GetLatestCommitSHAInBranch(ctx context.Context, owner string, repoName string, branch string) (string, error) {
	ret := _m.Called(ctx, owner, repoName, branch)

	if len(ret) == 0 {
		panic("no return value specified for GetLatestCommitSHAInBranch")
	}

	var r0 string
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string, string) (string, error)); ok {
		return rf(ctx, owner, repoName, branch)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, string, string) string); ok {
		r0 = rf(ctx, owner, repoName, branch)
	} else {
		r0 = ret.Get(0).(string)
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, string, string) error); ok {
		r1 = rf(ctx, owner, repoName, branch)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetPRBranchSHAFromPRNumber provides a mock function with given fields: ctx, owner, repoName, prNumber
func (_m *SCM) GetPRBranchSHAFromPRNumber(ctx context.Context, owner string, repoName string, prNumber int) (*string, *string, error) {
	ret := _m.Called(ctx, owner, repoName, prNumber)

	if len(ret) == 0 {
		panic("no return value specified for GetPRBranchSHAFromPRNumber")
	}

	var r0 *string
	var r1 *string
	var r2 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string, int) (*string, *string, error)); ok {
		return rf(ctx, owner, repoName, prNumber)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, string, int) *string); ok {
		r0 = rf(ctx, owner, repoName, prNumber)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*string)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, string, int) *string); ok {
		r1 = rf(ctx, owner, repoName, prNumber)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(*string)
		}
	}

	if rf, ok := ret.Get(2).(func(context.Context, string, string, int) error); ok {
		r2 = rf(ctx, owner, repoName, prNumber)
	} else {
		r2 = ret.Error(2)
	}

	return r0, r1, r2
}

// GetPRComments provides a mock function with given fields: ctx, owner, repoName, prNumber
func (_m *SCM) GetPRComments(ctx context.Context, owner string, repoName string, prNumber int) ([]*github.IssueComment, error) {
	ret := _m.Called(ctx, owner, repoName, prNumber)

	if len(ret) == 0 {
		panic("no return value specified for GetPRComments")
	}

	var r0 []*github.IssueComment
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string, int) ([]*github.IssueComment, error)); ok {
		return rf(ctx, owner, repoName, prNumber)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, string, int) []*github.IssueComment); ok {
		r0 = rf(ctx, owner, repoName, prNumber)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*github.IssueComment)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, string, int) error); ok {
		r1 = rf(ctx, owner, repoName, prNumber)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetPRCommentsByUser provides a mock function with given fields: ctx, owner, repoName, prNumber
func (_m *SCM) GetPRCommentsByUser(ctx context.Context, owner string, repoName string, prNumber int) ([]*github.IssueComment, error) {
	ret := _m.Called(ctx, owner, repoName, prNumber)

	if len(ret) == 0 {
		panic("no return value specified for GetPRCommentsByUser")
	}

	var r0 []*github.IssueComment
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string, int) ([]*github.IssueComment, error)); ok {
		return rf(ctx, owner, repoName, prNumber)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, string, int) []*github.IssueComment); ok {
		r0 = rf(ctx, owner, repoName, prNumber)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*github.IssueComment)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, string, int) error); ok {
		r1 = rf(ctx, owner, repoName, prNumber)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetPullRequestBodyFromCommit provides a mock function with given fields: ctx, owner, repoName, sha
func (_m *SCM) GetPullRequestBodyFromCommit(ctx context.Context, owner string, repoName string, sha string) (*string, error) {
	ret := _m.Called(ctx, owner, repoName, sha)

	if len(ret) == 0 {
		panic("no return value specified for GetPullRequestBodyFromCommit")
	}

	var r0 *string
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string, string) (*string, error)); ok {
		return rf(ctx, owner, repoName, sha)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, string, string) *string); ok {
		r0 = rf(ctx, owner, repoName, sha)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*string)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, string, string) error); ok {
		r1 = rf(ctx, owner, repoName, sha)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetPullRequestBodyFromPRNumber provides a mock function with given fields: ctx, owner, repoName, prNumber
func (_m *SCM) GetPullRequestBodyFromPRNumber(ctx context.Context, owner string, repoName string, prNumber int) (*string, error) {
	ret := _m.Called(ctx, owner, repoName, prNumber)

	if len(ret) == 0 {
		panic("no return value specified for GetPullRequestBodyFromPRNumber")
	}

	var r0 *string
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string, int) (*string, error)); ok {
		return rf(ctx, owner, repoName, prNumber)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, string, int) *string); ok {
		r0 = rf(ctx, owner, repoName, prNumber)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*string)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, string, int) error); ok {
		r1 = rf(ctx, owner, repoName, prNumber)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// HandleError provides a mock function with given fields: ctx, statusContext, owner, repoName, prNumber, headSHA, prOwner, err
func (_m *SCM) HandleError(ctx context.Context, statusContext string, owner string, repoName string, prNumber int, headSHA string, prOwner string, err error) error {
	ret := _m.Called(ctx, statusContext, owner, repoName, prNumber, headSHA, prOwner, err)

	if len(ret) == 0 {
		panic("no return value specified for HandleError")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string, string, int, string, string, error) error); ok {
		r0 = rf(ctx, statusContext, owner, repoName, prNumber, headSHA, prOwner, err)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// NewSCM creates a new instance of SCM. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewSCM(t interface {
	mock.TestingT
	Cleanup(func())
}) *SCM {
	mock := &SCM{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
