package github

import (
	"context"
	"fmt"
	"github.com/google/go-github/v48/github"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
	"sort"
)

type Client struct {
	github   *github.Client
	user     string
	owner    string
	repo     string
	prNumber int
}

func NewClient(owner, repo, user, token string, prNumber int) *Client {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)
	return &Client{
		github:   github.NewClient(tc),
		user:     user,
		owner:    owner,
		repo:     repo,
		prNumber: prNumber,
	}
}

func (c *Client) GetPullRequestBodyFromCommit(ctx context.Context, sha string) (*string, error) {
	prsWithCommit, _, err := c.github.PullRequests.ListPullRequestsWithCommit(ctx, c.owner, c.repo, sha, nil)
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

func (c *Client) GetPullRequestBodyFromPRNumber(ctx context.Context) (*string, error) {
	pr, _, err := c.github.PullRequests.Get(ctx, c.owner, c.repo, c.prNumber)
	if err != nil {
		return nil, err
	}
	return pr.Body, nil
}

func (c *Client) CommentOnPR(ctx context.Context, body string) error {
	_, _, err := c.github.Issues.CreateComment(ctx, c.owner, c.repo, c.prNumber, &github.IssueComment{Body: &body})
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) CommentError(ctx context.Context, err error) error {
	errorMsg := fmt.Sprintf("Validation Failed:\n%s", err.Error())
	return c.CommentOnPR(ctx, errorMsg)
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

func (c *Client) GetPRComments(ctx context.Context) ([]*github.IssueComment, error) {
	comments, _, err := c.github.Issues.ListComments(ctx, c.owner, c.repo, c.prNumber, nil)
	if err != nil {
		return nil, err
	}
	// Issue comments are returned sorted by ID, but we want them sorted by created time
	sort.Slice(comments, func(i, j int) bool {
		return comments[i].CreatedAt.After(*comments[j].CreatedAt)
	})
	return comments, nil
}

func (c *Client) GetPRCommentsByUser(ctx context.Context) ([]*github.IssueComment, error) {
	comments, err := c.GetPRComments(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get all PR comments")
	}
	var userComments []*github.IssueComment
	for _, comment := range comments {
		if *comment.User.Login == c.user {
			userComments = append(userComments, comment)
		}
	}
	return comments, nil
}

func (c *Client) GetFileFromBranch(ctx context.Context, branch, path string) ([]byte, error) {
	resp, _, _, err := c.github.Repositories.GetContents(ctx, c.owner, c.repo, path, &github.RepositoryContentGetOptions{Ref: branch})
	if err != nil {
		return nil, err
	}
	content, err := resp.GetContent()
	if err != nil {
		return nil, err
	}
	return []byte(content), nil
}

func (c *Client) DeleteUsersComments(ctx context.Context) error {
	comments, err := c.GetPRCommentsByUser(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get all PR comments")
	}
	for _, comment := range comments {
		_, err = c.github.Issues.DeleteComment(ctx, c.owner, c.repo, *comment.ID)
		if err != nil {
			return errors.Wrap(err, "failed to delete comment")
		}
	}
	return nil
}

func (c *Client) CreateCommitStatus(ctx context.Context, ref string, status *github.RepoStatus) error {
	_, _, err := c.github.Repositories.CreateStatus(ctx, c.owner, c.repo, ref, status)
	if err != nil {
		return errors.Wrap(err, "failed to create commit status")
	}
	return nil
}
