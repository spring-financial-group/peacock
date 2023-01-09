package github

import (
	"context"
	"fmt"
	"github.com/google/go-github/v47/github"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
	"sort"
)

type Client struct {
	Github *github.Client
}

func NewClient(token string) *Client {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)
	return &Client{
		Github: github.NewClient(tc),
	}
}

func (c *Client) GetPullRequestBodyFromCommit(ctx context.Context, owner, repo, sha string) (*string, error) {
	prsWithCommit, _, err := c.Github.PullRequests.ListPullRequestsWithCommit(ctx, owner, repo, sha, nil)
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

func (c *Client) GetPullRequestBodyFromPRNumber(ctx context.Context, owner, repo string, prNumber int) (*string, error) {
	pr, _, err := c.Github.PullRequests.Get(ctx, owner, repo, prNumber)
	if err != nil {
		return nil, err
	}
	return pr.Body, nil
}

func (c *Client) CommentOnPR(ctx context.Context, owner, repo string, prNumber int, body string) error {
	_, _, err := c.Github.Issues.CreateComment(ctx, owner, repo, prNumber, &github.IssueComment{Body: &body})
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) CommentError(ctx context.Context, owner, repo string, prNumber int, err error) error {
	errorMsg := fmt.Sprintf("Validation Failed:\n%s", err.Error())
	return c.CommentOnPR(ctx, owner, repo, prNumber, errorMsg)
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

func (c *Client) GetPRComments(ctx context.Context, owner, repo string, prNumber int) ([]*github.IssueComment, error) {
	comments, _, err := c.Github.Issues.ListComments(ctx, owner, repo, prNumber, nil)
	if err != nil {
		return nil, err
	}
	// Issue comments are returned sorted by ID, but we want them sorted by created time
	sort.Slice(comments, func(i, j int) bool {
		return comments[i].CreatedAt.After(*comments[j].CreatedAt)
	})
	return comments, nil
}

func (c *Client) GetPRCommentsByUser(ctx context.Context, owner, repo, user string, prNumber int) ([]*github.IssueComment, error) {
	comments, err := c.GetPRComments(ctx, owner, repo, prNumber)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get all PR comments")
	}
	var userComments []*github.IssueComment
	for _, comment := range comments {
		if *comment.User.Login == user {
			userComments = append(userComments, comment)
		}
	}
	return comments, nil
}

func (c *Client) GetFileFromBranch(ctx context.Context, owner, repo, branch, path string) ([]byte, error) {
	resp, _, _, err := c.Github.Repositories.GetContents(ctx, owner, repo, path, &github.RepositoryContentGetOptions{Ref: branch})
	if err != nil {
		return nil, err
	}
	content, err := resp.GetContent()
	if err != nil {
		return nil, err
	}
	return []byte(content), nil
}
