package domain

import (
	"context"
	"github.com/google/go-github/v48/github"
	githubscm "github.com/spring-financial-group/peacock/pkg/git/github"
)

const (
	GitHubURL = "https://github.com"
)

// Repository status constants
const (
	SuccessStatus = "success"
	PendingStatus = "pending"
	FailureStatus = "failure"
	ErrorStatus   = "error"
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
	// CreateCommitStatus creates a commit status on a commit
	CreateCommitStatus(ctx context.Context, ref string, status *github.RepoStatus) error
	// CreateValidationCommitStatus creates a validation commit status on a commit
	CreateValidationCommitStatus(ctx context.Context, ref string, state string) error
	// CreateReleaseCommitStatus creates a release commit status on a commit
	CreateReleaseCommitStatus(ctx context.Context, ref string, state string) error
}

type SCMClientFactory interface {
	// GetClient returns a client for interacting with the SCM
	GetClient(owner, repo, user string, prNumber int) *githubscm.Client
	// RemoveClient removes a client from memory
	RemoveClient(client *githubscm.Client)
}
