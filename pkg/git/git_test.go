package git

import (
	"context"
	"fmt"
	"github.com/google/go-github/v47/github"
	"github.com/jenkins-x/jx-helpers/v3/pkg/gitclient/giturl"
	ghmock "github.com/migueleliasweb/go-github-mock/src/mock"
	"github.com/spring-financial-group/peacock/pkg/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
	"time"
)

type mockGitter struct {
	mock.Mock
	returnedSHA string
}

func (m *mockGitter) Command(dir string, args ...string) (string, error) {
	mArgs := m.Called(dir, args)
	return mArgs.String(0), mArgs.Error(1)
}

func TestGit_GetPullRequestFromLastCommit(t *testing.T) {
	onePM, _ := time.Parse("15:04", "13:00")
	twoPM, _ := time.Parse("15:04", "14:00")
	threePM, _ := time.Parse("15:04", "15:00")
	fourPM, _ := time.Parse("15:04", "16:00")

	testCases := []struct {
		name        string
		returnedPRs []*github.PullRequest
		expectedPR  *github.PullRequest
		shouldError bool
	}{
		{
			name: "OnePRFound",
			expectedPR: &github.PullRequest{
				Merged:   utils.NewPtr(true),
				MergedAt: &onePM,
			},
			returnedPRs: []*github.PullRequest{
				{
					Merged:   utils.NewPtr(true),
					MergedAt: &onePM,
				},
			},
			shouldError: false,
		},
		{
			name:        "NoPRFound",
			expectedPR:  nil,
			returnedPRs: nil,
			shouldError: true,
		},
		{
			name: "ManyFound",
			expectedPR: &github.PullRequest{
				Merged:   utils.NewPtr(true),
				MergedAt: &onePM,
			},
			returnedPRs: []*github.PullRequest{
				{
					Merged:   utils.NewPtr(true),
					MergedAt: &twoPM,
				},
				{
					Merged:   utils.NewPtr(true),
					MergedAt: &fourPM,
				},
				{
					Merged:   utils.NewPtr(true),
					MergedAt: &onePM,
				},
				{
					Merged:   utils.NewPtr(true),
					MergedAt: &threePM,
				},
			},
			shouldError: false,
		},
		{
			name: "MoreRecentPrByIsNotMerged",
			expectedPR: &github.PullRequest{
				Merged:   utils.NewPtr(true),
				MergedAt: &twoPM,
			},
			returnedPRs: []*github.PullRequest{
				{
					Merged:   utils.NewPtr(true),
					MergedAt: &twoPM,
				},
				{
					Merged:   utils.NewPtr(true),
					MergedAt: &fourPM,
				},
				{
					Merged:   utils.NewPtr(false),
					MergedAt: &onePM,
				},
				{
					Merged:   utils.NewPtr(true),
					MergedAt: &threePM,
				},
			},
			shouldError: false,
		},
	}

	for _, tt := range testCases {
		mockGitter := new(mockGitter)
		mockedHTTPClient := ghmock.NewMockedHTTPClient(
			ghmock.WithRequestMatch(
				ghmock.GetReposCommitsPullsByOwnerByRepoByCommitSha,
				tt.returnedPRs,
			),
		)
		mockGH := github.NewClient(mockedHTTPClient)

		mockGitter.On("Command", "", []string{"rev-parse", "HEAD"}).Return("testSHA", nil)

		client := &Client{
			github: mockGH,
			gitter: mockGitter,
			owner:  "peacock",
			repo:   "repo",
		}

		t.Run(tt.name, func(t *testing.T) {
			actualPR, err := client.GetPullRequestFromLastCommit(context.Background())
			if tt.shouldError {
				fmt.Println("expected error: " + err.Error())
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.expectedPR, actualPR)
		})
	}
}

func TestGit_NewClient(t *testing.T) {
	type inputArgs struct {
		serverURL string
		owner     string
		repo      string
		token     string
	}

	testCases := []struct {
		name string
		inputArgs
		shouldError bool
	}{
		{
			name: "Passing",
			inputArgs: inputArgs{
				serverURL: giturl.GitHubURL,
				owner:     "owner",
				repo:      "peacock",
				token:     "token",
			},
			shouldError: false,
		},
		{
			name: "NonGithubURL",
			inputArgs: inputArgs{
				serverURL: giturl.FakeGitURL,
				owner:     "owner",
				repo:      "peacock",
				token:     "token",
			},
			shouldError: true,
		},
		{
			name: "NoOwner",
			inputArgs: inputArgs{
				serverURL: giturl.GitHubURL,
				repo:      "peacock",
				token:     "token",
			},
			shouldError: false,
		},
		{
			name: "NoRepo",
			inputArgs: inputArgs{
				serverURL: giturl.GitHubURL,
				owner:     "owner",
				token:     "token",
			},
			shouldError: false,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewClient(tt.serverURL, tt.owner, tt.repo, tt.token)
			if tt.shouldError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
