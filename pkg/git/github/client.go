package github

import (
	"context"
	"fmt"
	"github.com/google/go-github/v48/github"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spring-financial-group/peacock/pkg/domain"
	"github.com/spring-financial-group/peacock/pkg/utils"
	"golang.org/x/oauth2"
	"net/http"
	"sort"
)

const (
	ValidationContext = "peacock-validation"
	ReleaseContext    = "peacock-release"
)

// RepoStatus Base repository statuses to use for creating a commit status
var (
	RepoStatus = map[string]*github.RepoStatus{
		ValidationContext: {
			State:       nil,
			Description: utils.NewPtr("Validates the PR body against the feathers"),
			Context:     utils.NewPtr(ValidationContext),
		},
		ReleaseContext: {
			State:       nil,
			Description: utils.NewPtr("Sends the messages to the Teams outlined in the PR body"),
			Context:     utils.NewPtr(ReleaseContext),
		},
	}
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
	fileContent, _, resp, err := c.github.Repositories.GetContents(ctx, c.owner, c.repo, path, &github.RepositoryContentGetOptions{Ref: branch})
	if err != nil {
		if resp.StatusCode == http.StatusNotFound {
			err = &domain.ErrFileNotFound{Path: path}
		}
		return nil, errors.Wrap(err, "failed to get file from branch")
	}

	content, err := fileContent.GetContent()
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

func (c *Client) CreatePeacockCommitStatus(ctx context.Context, ref string, state domain.State, statusContext string) error {
	status := RepoStatus[statusContext]
	status.State = utils.NewPtr(string(state))

	_, _, err := c.github.Repositories.CreateStatus(ctx, c.owner, c.repo, ref, status)
	if err != nil {
		return errors.Wrap(err, "failed to create commit status")
	}
	return nil
}

func (c *Client) GetLatestCommitSHAInBranch(ctx context.Context, branch string) (string, error) {
	commit, _, err := c.github.Repositories.GetCommit(ctx, c.owner, c.repo, branch, nil)
	if err != nil {
		return "", errors.Wrap(err, "failed to get latest commit in branch")
	}
	return commit.GetSHA(), nil
}

func (c *Client) GetKey() string {
	return fmt.Sprintf("%s/%s/%s/%d", c.user, c.owner, c.repo, c.prNumber)
}

func (c *Client) HandleError(ctx context.Context, statusContext, headSHA string, err error) error {
	commentErr := c.CommentError(ctx, err)
	if commentErr != nil {
		log.Errorf("Failed to comment error on PR: %s", commentErr.Error())
	}

	statusErr := c.CreatePeacockCommitStatus(ctx, headSHA, domain.FailureState, statusContext)
	if statusErr != nil {
		log.Errorf("failed to create failed commit status: %s", err)
	}
	return err
}
