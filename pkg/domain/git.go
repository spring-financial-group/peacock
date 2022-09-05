package domain

import (
	"context"
	"github.com/google/go-github/v47/github"
)

type Git interface {
	// GetPullRequest returns a github.PullRequest from pr number
	GetPullRequest(ctx context.Context, prNumber int) (*github.PullRequest, error)
	// CommentOnPR posts a comment to a given pull request
	CommentOnPR(ctx context.Context, pullRequest *github.PullRequest, body string) error
}
