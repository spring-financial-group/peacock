package github

import (
	"context"
	"github.com/google/go-github/v47/github"
	"github.com/pkg/errors"
	"github.com/spring-financial-group/mqa-logging/pkg/log"
	"github.com/spring-financial-group/peacock/pkg/domain"
	"golang.org/x/oauth2"
	"net/http"
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
	prsWithCommit, r, err := c.Github.PullRequests.ListPullRequestsWithCommit(ctx, c.Owner, c.Repo, sha, nil)
	if err != nil {
		return nil, err
	}
	if r.StatusCode != http.StatusOK {
		return nil, errors.Errorf("failed to get pull requests, %d code recieved", r.StatusCode)
	}
	if len(prsWithCommit) < 1 {
		return nil, errors.New("no pull request found containing commit")
	}
	log.Logger().Infof("Found %d pull request(s) containing that commit", len(prsWithCommit))

	// If there is only one PR then that must be is
	if len(prsWithCommit) == 1 {
		return prsWithCommit[0].Body, nil
	}
	return c.findPRByMergedTime(prsWithCommit).Body, nil
}

func (c *Client) GetPullRequestBodyFromPRNumber(ctx context.Context, prNumber int) (*string, error) {
	pr, resp, err := c.Github.PullRequests.Get(ctx, c.Owner, c.Repo, prNumber)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, errors.Errorf("failed to get pull request %d code recieved", resp.StatusCode)
	}
	return pr.Body, nil
}

func (c *Client) CommentOnPR(ctx context.Context, prNumber int, body string) error {
	_, r, err := c.Github.Issues.CreateComment(ctx, c.Owner, c.Repo, prNumber, &github.IssueComment{Body: &body})
	if err != nil {
		return err
	}
	if r.StatusCode != http.StatusCreated {
		return errors.Errorf("github response code %d", r.StatusCode)
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
