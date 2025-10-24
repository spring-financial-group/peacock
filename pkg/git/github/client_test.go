package github

import (
	"context"
	"github.com/google/go-github/v48/github"
	ghmock "github.com/migueleliasweb/go-github-mock/src/mock"
	"github.com/spring-financial-group/peacock/pkg/utils"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestGit_GetPullRequestBodyWithCommit(t *testing.T) {
	onePM, _ := time.Parse("15:04", "13:00")
	twoPM, _ := time.Parse("15:04", "14:00")
	threePM, _ := time.Parse("15:04", "15:00")
	fourPM, _ := time.Parse("15:04", "16:00")

	testCases := []struct {
		name         string
		returnedPRs  []*github.PullRequest
		expectedBody *string
		shouldError  bool
	}{
		{
			name: "OnePRFound",
			returnedPRs: []*github.PullRequest{
				{
					Body:     utils.NewPtr("PR1"),
					Merged:   utils.NewPtr(true),
					MergedAt: &onePM,
				},
			},
			expectedBody: utils.NewPtr("PR1"),
			shouldError:  false,
		},
		{
			name:         "NoPRFound",
			expectedBody: nil,
			returnedPRs:  nil,
			shouldError:  true,
		},
		{
			name: "ManyFound",
			returnedPRs: []*github.PullRequest{
				{
					Body:     utils.NewPtr("PR1"),
					Merged:   utils.NewPtr(true),
					MergedAt: &twoPM,
				},
				{
					Body:     utils.NewPtr("PR2"),
					Merged:   utils.NewPtr(true),
					MergedAt: &fourPM,
				},
				{
					Body:     utils.NewPtr("PR3"),
					Merged:   utils.NewPtr(true),
					MergedAt: &onePM,
				},
				{
					Body:     utils.NewPtr("PR4"),
					Merged:   utils.NewPtr(true),
					MergedAt: &threePM,
				},
			},
			expectedBody: utils.NewPtr("PR3"),
			shouldError:  false,
		},
		{
			name: "MoreRecentPrByIsNotMerged",
			returnedPRs: []*github.PullRequest{
				{
					Body:     utils.NewPtr("PR1"),
					Merged:   utils.NewPtr(true),
					MergedAt: &twoPM,
				},
				{
					Body:     utils.NewPtr("PR2"),
					Merged:   utils.NewPtr(true),
					MergedAt: &fourPM,
				},
				{
					Body:     utils.NewPtr("PR3"),
					Merged:   utils.NewPtr(false),
					MergedAt: &onePM,
				},
				{
					Body:     utils.NewPtr("PR4"),
					Merged:   utils.NewPtr(true),
					MergedAt: &threePM,
				},
			},
			expectedBody: utils.NewPtr("PR1"),
			shouldError:  false,
		},
	}

	for _, tt := range testCases {
		mockedHTTPClient := ghmock.NewMockedHTTPClient(
			ghmock.WithRequestMatch(
				ghmock.GetReposCommitsPullsByOwnerByRepoByCommitSha,
				tt.returnedPRs,
			),
		)
		mockGH := github.NewClient(mockedHTTPClient)

		client := Client{
			github: mockGH,
			user:   "mqube-bot",
		}

		t.Run(tt.name, func(t *testing.T) {
			actualBody, err := client.GetPullRequestBodyFromCommit(context.Background(), "spring-financial-group", "peacock", "CommitSHA")
			if tt.shouldError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.expectedBody, actualBody)
		})
	}
}
