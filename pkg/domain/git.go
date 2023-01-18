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
	GetPullRequestBodyFromCommit(ctx context.Context, sha string) (*string, error)
	// GetPullRequestBodyFromPRNumber returns the body of a pull request from pr number
	GetPullRequestBodyFromPRNumber(ctx context.Context) (*string, error)
	// CommentOnPR posts a comment on a pull request given the pr number
	CommentOnPR(ctx context.Context, body string) error
	// CommentError posts an error comment on a pull request given the pr number
	CommentError(ctx context.Context, err error) error
	// GetPRComments returns all comments on a pull request given the pr number sorted by most recent comment first
	GetPRComments(ctx context.Context) ([]*github.IssueComment, error)
	// GetFileFromBranch returns the file as a string from a branch
	GetFileFromBranch(ctx context.Context, branch, path string) ([]byte, error)
	// GetPRCommentsByUser returns all the comments on a pull request by a user
	GetPRCommentsByUser(ctx context.Context) ([]*github.IssueComment, error)
	// DeleteUsersComments deletes all the comments on a pull request by a user
	DeleteUsersComments(ctx context.Context) error
	// CreatePeacockCommitStatus creates a commit status on a commit
	CreatePeacockCommitStatus(ctx context.Context, ref string, state State, statusContext string) error
	// GetLatestCommitSHAInBranch returns the most recent commit in a branch
	GetLatestCommitSHAInBranch(ctx context.Context, branch string) (string, error)
	// GetKey returns the identifier of the SCM
	GetKey() string
	// HandleError handles an error by commenting on the PR and creating a commit status on the given SHA
	HandleError(ctx context.Context, statusContext, headSHA string, err error) error
}

type SCMClientFactory interface {
	// GetClient returns a client for interacting with the SCM
	GetClient(owner, repo, user string, prNumber int) SCM
	// RemoveClient removes a client from memory
	RemoveClient(key string)
}
