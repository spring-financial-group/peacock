package domain

import (
	"context"
	"github.com/google/go-github/v48/github"
)

const (
	GitHubURL = "https://github.com"
)

// Repository status constants
const (
	SuccessState = State("success")
	PendingState = State("pending")
	FailureState = State("failure")
	ErrorState   = State("error")
)

// State is a type alias for the state of a repository status
type State string

// Repository status contexts
const (
	ValidationContext = "peacock-validation"
	ReleaseContext    = "peacock-release"
)

type Git interface {
	// GetLatestCommitSHA gets the SHA of the latest commit from the local env
	GetLatestCommitSHA(dir string) (string, error)
	// GetRepoOwnerAndName gets the owner and name of the repos from the local env
	GetRepoOwnerAndName(dir string) (string, string, error)
}

type SCM interface {
	GetPullRequestBodyFromCommit(ctx context.Context, owner, repoName, sha string) (*string, error)
	// GetPullRequestBodyFromPRNumber returns the body of a pull request from pr number
	GetPullRequestBodyFromPRNumber(ctx context.Context, owner, repoName string, prNumber int) (*string, error)
	// GetPRBranchSHAFromPRNumber returns the Branch and Head SHA of a pull request from pr number
	GetPRBranchSHAFromPRNumber(ctx context.Context, owner, repoName string, prNumber int) (*string, *string, error)
	// CommentOnPR posts a comment on a pull request given the pr number
	CommentOnPR(ctx context.Context, owner, repoName string, prNumber int, body string) error
	// CommentError posts an error comment on a pull request given the pr number
	CommentError(ctx context.Context, owner, repoName string, prNumber int, prOwner string, err error) error
	// GetPRComments returns all comments on a pull request given the pr number sorted by most recent comment first
	GetPRComments(ctx context.Context, owner, repoName string, prNumber int) ([]*github.IssueComment, error)
	// GetFileFromBranch returns the file as a string from a branch
	GetFileFromBranch(ctx context.Context, owner, repoName, branch, path string) ([]byte, error)
	// GetPRCommentsByUser returns all the comments on a pull request by a user
	GetPRCommentsByUser(ctx context.Context, owner, repoName string, prNumber int) ([]*github.IssueComment, error)
	// DeleteUsersComments deletes all the comments on a pull request by a user
	DeleteUsersComments(ctx context.Context, owner, repoName string, prNumber int) error
	// CreatePeacockCommitStatus creates a commit status on a commit
	CreatePeacockCommitStatus(ctx context.Context, owner, repoName, ref string, state State, statusContext string) error
	// GetLatestCommitSHAInBranch returns the most recent commit in a branch
	GetLatestCommitSHAInBranch(ctx context.Context, owner, repoName, branch string) (string, error)
	// HandleError handles an error by commenting on the PR and creating a commit status on the given SHA
	HandleError(ctx context.Context, statusContext, owner, repoName string, prNumber int, headSHA, prOwner string, err error) error
	// GetFilesChangedFromPR returns the files changed files in the given pr
	GetFilesChangedFromPR(ctx context.Context, owner string, repoName string, prNumber int) ([]*github.CommitFile, error)
}
