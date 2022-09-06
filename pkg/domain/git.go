package domain

import (
	"context"
	"github.com/google/go-github/v47/github"
)

type Git interface {
	GetPullRequestFromLastCommit(ctx context.Context) (*github.PullRequest, error)
	// GetPullRequestFromPRNumber returns a github.PullRequest from pr number
	GetPullRequestFromPRNumber(ctx context.Context, prNumber int) (*github.PullRequest, error)
	// CommentOnPR posts a comment to a given pull request
	CommentOnPR(ctx context.Context, pullRequest *github.PullRequest, body string) error
}
