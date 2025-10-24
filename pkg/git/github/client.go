package github

import (
	"context"
	"fmt"
	"github.com/google/go-github/v48/github"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
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
	github *github.Client
	user   string
}

func NewClient(user, token string) *Client {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)
	return &Client{
		github: github.NewClient(tc),
		user:   user,
	}
}

func (c *Client) GetPullRequestBodyFromCommit(ctx context.Context, owner, repoName, sha string) (*string, error) {
	prsWithCommit, _, err := c.github.PullRequests.ListPullRequestsWithCommit(ctx, owner, repoName, sha, nil)
	if err != nil {
		return nil, err
	}
	if len(prsWithCommit) < 1 {
		return nil, errors.Errorf("no pull request found containing commit %s", sha)
	}
	log.Info().Msgf("Found %d pull request(s) containing that commit", len(prsWithCommit))

	// If there is only one PR then that must be it
	if len(prsWithCommit) == 1 {
		return prsWithCommit[0].Body, nil
	}
	return c.findPRByMergedTime(prsWithCommit).Body, nil
}

func (c *Client) GetPullRequestBodyFromPRNumber(ctx context.Context, owner, repoName string, prNumber int) (*string, error) {
	pr, _, err := c.github.PullRequests.Get(ctx, owner, repoName, prNumber)
	if err != nil {
		return nil, err
	}
	return pr.Body, nil
}

func (c *Client) GetPRBranchSHAFromPRNumber(ctx context.Context, owner, repoName string, prNumber int) (*string, *string, error) {
	pr, _, err := c.github.PullRequests.Get(ctx, owner, repoName, prNumber)
	if err != nil {
		return nil, nil, err
	}
	return pr.Head.Ref, pr.Head.SHA, nil
}

func (c *Client) CommentOnPR(ctx context.Context, owner, repoName string, prNumber int, body string) error {
	_, _, err := c.github.Issues.CreateComment(ctx, owner, repoName, prNumber, &github.IssueComment{Body: &body})
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) CommentError(ctx context.Context, owner, repoName string, prNumber int, prOwner string, err error) error {
	var tagString string
	if prOwner != "" {
		tagString = fmt.Sprintf("@%s: ", prOwner)
	}
	errorMsg := fmt.Sprintf("%sValidation failed for the release notes in this PR:\n%s", tagString, err.Error())
	return c.CommentOnPR(ctx, owner, repoName, prNumber, errorMsg)
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

func (c *Client) GetPRComments(ctx context.Context, owner, repoName string, prNumber int) ([]*github.IssueComment, error) {
	comments, _, err := c.github.Issues.ListComments(ctx, owner, repoName, prNumber, nil)
	if err != nil {
		return nil, err
	}
	// Issue comments are returned sorted by ID, but we want them sorted by created time
	sort.Slice(comments, func(i, j int) bool {
		return comments[i].CreatedAt.After(*comments[j].CreatedAt)
	})
	return comments, nil
}

func (c *Client) GetPRCommentsByUser(ctx context.Context, owner, repoName string, prNumber int) ([]*github.IssueComment, error) {
	comments, err := c.GetPRComments(ctx, owner, repoName, prNumber)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get all PR comments")
	}
	var userComments []*github.IssueComment
	for _, comment := range comments {
		if *comment.User.Login == c.user {
			userComments = append(userComments, comment)
		}
	}
	return userComments, nil
}

func (c *Client) GetFileFromBranch(ctx context.Context, owner, repoName, branch, path string) ([]byte, error) {
	fileContent, _, resp, err := c.github.Repositories.GetContents(ctx, owner, repoName, path, &github.RepositoryContentGetOptions{Ref: branch})
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

func (c *Client) GetFilesChangedFromPR(ctx context.Context, owner string, repoName string, prNumber int) ([]*github.CommitFile, error) {
	commitFiles, resp, err := c.github.PullRequests.ListFiles(ctx, owner, repoName, prNumber, nil)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, errors.New("failed to retrieve changed files from pull request")
	}

	return commitFiles, nil
}

func (c *Client) DeleteUsersComments(ctx context.Context, owner, repoName string, prNumber int) error {
	comments, err := c.GetPRCommentsByUser(ctx, owner, repoName, prNumber)
	if err != nil {
		return errors.Wrap(err, "failed to get all PR comments")
	}
	for _, comment := range comments {
		_, err = c.github.Issues.DeleteComment(ctx, owner, repoName, *comment.ID)
		if err != nil {
			return errors.Wrap(err, "failed to delete comment")
		}
	}
	return nil
}

func (c *Client) CreatePeacockCommitStatus(ctx context.Context, owner, repoName, ref string, state domain.State, statusContext string) error {
	status := RepoStatus[statusContext]
	status.State = utils.NewPtr(string(state))

	_, _, err := c.github.Repositories.CreateStatus(ctx, owner, repoName, ref, status)
	if err != nil {
		return errors.Wrap(err, "failed to create commit status")
	}
	return nil
}

func (c *Client) GetLatestCommitSHAInBranch(ctx context.Context, owner, repoName, branch string) (string, error) {
	commit, _, err := c.github.Repositories.GetCommit(ctx, owner, repoName, branch, nil)
	if err != nil {
		return "", errors.Wrap(err, "failed to get latest commit in branch")
	}
	return commit.GetSHA(), nil
}

func (c *Client) HandleError(ctx context.Context, statusContext, owner, repoName string, prNumber int, headSHA, prOwner string, err error) error {
	commentErr := c.CommentError(ctx, owner, repoName, prNumber, prOwner, err)
	if commentErr != nil {
		log.Error().Msgf("Failed to comment error on PR: %s", commentErr.Error())
	}

	statusErr := c.CreatePeacockCommitStatus(ctx, owner, repoName, headSHA, domain.FailureState, statusContext)
	if statusErr != nil {
		log.Error().Msgf("failed to create failed commit status: %s", err)
	}
	return err
}
