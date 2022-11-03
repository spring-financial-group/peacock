package github

import (
	"context"
	"github.com/google/go-github/v47/github"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spring-financial-group/peacock/pkg/domain"
	"golang.org/x/oauth2"
)

type Client struct {
	Github *github.Client

	Owner string
	Repo  string
}

func NewClient(owner, repo, token string) domain.GitServer {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)
	return &Client{
		Github: github.NewClient(tc),
		Owner:  owner,
		Repo:   repo,
	}
}

func (c *Client) GetPullRequestBodyFromCommit(ctx context.Context, sha string) (*string, error) {
	prsWithCommit, _, err := c.Github.PullRequests.ListPullRequestsWithCommit(ctx, c.Owner, c.Repo, sha, nil)
	if err != nil {
		return nil, err
	}
	if len(prsWithCommit) < 1 {
		return nil, errors.Errorf("no pull request found containing commit %s", sha)
	}
	log.Infof("Found %d pull request(s) containing that commit", len(prsWithCommit))

	// If there is only one PR then that must be is
	if len(prsWithCommit) == 1 {
		return prsWithCommit[0].Body, nil
	}
	return c.findPRByMergedTime(prsWithCommit).Body, nil
}

func (c *Client) GetPullRequestBodyFromPRNumber(ctx context.Context, prNumber int) (*string, error) {
	pr, _, err := c.Github.PullRequests.Get(ctx, c.Owner, c.Repo, prNumber)
	if err != nil {
		return nil, err
	}
	return pr.Body, nil
}

func (c *Client) CommentOnPR(ctx context.Context, prNumber int, body string) error {
	_, _, err := c.Github.Issues.CreateComment(ctx, c.Owner, c.Repo, prNumber, &github.IssueComment{Body: &body})
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) findPRByMergedTime(pullRequests []*github.PullRequest) *github.PullRequest {
	var mostRecentPR int
	for idx, pr := range pullRequests {
		// The commit must have come from a merged PR
		if !*pr.Merged {
			continue
		}

		// We assume that it's the most recent PR that caused the release
		if pr.MergedAt.Before(*pullRequests[mostRecentPR].MergedAt) {
			mostRecentPR = idx
		}
	}
	return pullRequests[mostRecentPR]
}

func (c *Client) GetPRComments(ctx context.Context, prNumber int) ([]*github.PullRequestComment, error) {
	comments, _, err := c.Github.PullRequests.ListComments(ctx, c.Owner, c.Repo, prNumber, &github.PullRequestListCommentsOptions{Sort: "created_at", Direction: "desc"})
	if err != nil {
		return nil, err
	}
	return comments, nil
}
