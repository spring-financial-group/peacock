package domain

import (
	"context"
)

const (
	GitHubURL = "https://github.com"
)

type Git interface {
	// GetLatestCommitSHA gets the SHA of the latest commit from the local env
	GetLatestCommitSHA() (string, error)
	// GetRepoOwnerAndName gets the owner and name of the repos from the local env
	GetRepoOwnerAndName() (string, string, error)
}

type GitServer interface {
	GetPullRequestBodyFromCommit(ctx context.Context, sha string) (*string, error)
	// GetPullRequestBodyFromPRNumber returns the body of a pull request from pr number
	GetPullRequestBodyFromPRNumber(ctx context.Context, prNumber int) (*string, error)
	// CommentOnPR posts a comment on a pull request given the pr number
	CommentOnPR(ctx context.Context, prNumber int, body string) error
}
