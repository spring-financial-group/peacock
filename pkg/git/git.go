package git

import (
	"context"
	"github.com/google/go-github/v47/github"
	"github.com/jenkins-x/jx-helpers/v3/pkg/gitclient"
	"github.com/jenkins-x/jx-helpers/v3/pkg/gitclient/cli"
	"github.com/jenkins-x/jx-helpers/v3/pkg/gitclient/giturl"
	"github.com/jenkins-x/jx-helpers/v3/pkg/scmhelpers"
	"github.com/jenkins-x/jx-logging/v3/pkg/log"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"
	"net/http"
	"net/url"
	"strings"
)

type Client struct {
	factory *scmhelpers.Factory
	github  *github.Client
	gitter  gitclient.Interface

	owner string
	repo  string
}

func NewClient(gitServerUrl, owner, repo, token string) (*Client, error) {
	w := new(Client)
	err := w.initClients(gitServerUrl, token)
	if err != nil {
		return nil, err
	}
	if owner == "" || repo == "" {
		log.Logger().Info("No owner or repo names provided, getting info from local instance")
		w.owner, w.repo, err = w.getOwnerAndRepo()
		if err != nil {
			return nil, errors.Wrap(err, "failed to get owner & repo name")
		}
	} else {
		w.owner = owner
		w.repo = repo
	}

	return w, nil
}

func (c *Client) GetPullRequestFromLastCommit(ctx context.Context) (*github.PullRequest, error) {
	latest, err := gitclient.GetLatestCommitSha(c.gitter, "")
	if err != nil {
		return nil, err
	}
	log.Logger().Infof("Found latest commit at %s", latest)

	prsWithCommit, r, err := c.github.PullRequests.ListPullRequestsWithCommit(ctx, c.owner, c.repo, latest, nil)
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

	// If there is only PR then that must be is
	if len(prsWithCommit) == 1 {
		return prsWithCommit[0], nil
	}

	return c.findPRByMergedTime(prsWithCommit), nil
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

func (c *Client) GetPullRequestFromPRNumber(ctx context.Context, prNumber int) (*github.PullRequest, error) {
	pr, resp, err := c.github.PullRequests.Get(ctx, c.owner, c.repo, prNumber)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, errors.Errorf("failed to get pull request %d code recieved", resp.StatusCode)
	}
	return pr, nil
}

func (c *Client) CommentOnPR(ctx context.Context, pullRequest *github.PullRequest, body string) error {
	comment := &github.IssueComment{
		Body:     &body,
		URL:      pullRequest.URL,
		HTMLURL:  pullRequest.HTMLURL,
		IssueURL: pullRequest.IssueURL,
	}
	_, r, err := c.github.Issues.CreateComment(ctx, c.owner, c.repo, *pullRequest.Number, comment)
	if err != nil {
		return err
	}
	if r.StatusCode != http.StatusCreated {
		return errors.Errorf("github response code %d", r.StatusCode)
	}
	return nil
}

func (c *Client) initClients(gitServerUrl, token string) error {
	c.factory = &scmhelpers.Factory{
		GitServerURL: gitServerUrl,
		GitToken:     token,
	}
	_, err := c.factory.Create()
	if err != nil {
		return err
	}
	if c.factory.GitKind != giturl.KindGitHub {
		return errors.Errorf("peacock doesn't currently support %s", c.factory.GitKind)
	}
	if c.github == nil {
		c.github = c.initGitHubClient(c.factory.GitToken)
	}
	if c.gitter == nil {
		c.gitter = cli.NewCLIClient("", nil)
	}
	return nil
}

func (c *Client) initGitHubClient(token string) *github.Client {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)
	return github.NewClient(tc)
}

func (c *Client) getOwnerAndRepo() (string, string, error) {
	command, err := c.gitter.Command("", "config", "--get", "remote.origin.url")
	if err != nil {
		return "", "", err
	}
	url, err := url.Parse(command)
	if err != nil {
		return "", "", err
	}
	path := strings.TrimSuffix(url.Path, ".git")
	path = strings.TrimPrefix(path, "/")
	split := strings.Split(path, "/")
	return split[0], split[1], nil
}
