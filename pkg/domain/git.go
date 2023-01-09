package domain

import (
	"context"
	"github.com/google/go-github/v47/github"
)

const (
	GitHubURL = "https://github.com"
)

type Git interface {
	// GetLatestCommitSHA gets the SHA of the latest commit from the local env
	GetLatestCommitSHA(dir string) (string, error)
	// GetRepoOwnerAndName gets the owner and name of the repos from the local env
	GetRepoOwnerAndName(dir string) (string, string, error)
}

type GitServer interface {
	GetPullRequestBodyFromCommit(ctx context.Context, owner, repo string, sha string) (*string, error)
	// GetPullRequestBodyFromPRNumber returns the body of a pull request from pr number
	GetPullRequestBodyFromPRNumber(ctx context.Context, owner, repo string, prNumber int) (*string, error)
	// CommentOnPR posts a comment on a pull request given the pr number
	CommentOnPR(ctx context.Context, owner, repo string, prNumber int, body string) error
	// CommentError posts an error comment on a pull request given the pr number
	CommentError(ctx context.Context, owner, repo string, prNumber int, err error) error
	// GetPRComments returns all comments on a pull request given the pr number sorted by most recent comment first
	GetPRComments(ctx context.Context, owner, repo string, prNumber int) ([]*github.IssueComment, error)
	// GetFileFromBranch returns the file as a string from a branch
	GetFileFromBranch(ctx context.Context, owner, repo, branch, path string) ([]byte, error)
	// GetPRCommentsByUser returns all the comments on a pull request by a user
	GetPRCommentsByUser(ctx context.Context, owner, repo, user string, prNumber int) ([]*github.IssueComment, error)
}
