package github

import (
	"context"
	"fmt"
	"github.com/google/go-github/v48/github"
	ghmock "github.com/migueleliasweb/go-github-mock/src/mock"
	"github.com/spring-financial-group/peacock/pkg/utils"
	"github.com/stretchr/testify/assert"
	"net/http"
	"strings"
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
				fmt.Println("expected error: " + err.Error())
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.expectedBody, actualBody)
		})
	}
}

func TestClient_DeletePeacockComments(t *testing.T) {
	now := time.Now()
	newComment := func(id int64, user, body string, age time.Duration) *github.IssueComment {
		return &github.IssueComment{
			ID:        utils.NewPtr(id),
			User:      &github.User{Login: utils.NewPtr(user)},
			Body:      utils.NewPtr(body),
			CreatedAt: utils.NewPtr(now.Add(-age)),
		}
	}

	var deleted []string
	mockedHTTPClient := ghmock.NewMockedHTTPClient(
		ghmock.WithRequestMatchPages(
			ghmock.GetReposIssuesCommentsByOwnerByRepoByIssueNumber,
			[]*github.IssueComment{
				newComment(1, "mqube-bot", "breakdown\n<!-- hash: abc123 type: peacock-breakdown -->\n", 5*time.Hour),
				newComment(2, "mqube-bot", "comment from another bot", 4*time.Hour),
				newComment(5, "mqube-bot", "other bot\n<!-- hash: abc123 type: summary -->\n", 4*time.Hour),
			},
			[]*github.IssueComment{
				newComment(3, "another-account", "breakdown\n<!-- hash: abc123 type: peacock-breakdown -->\n", 3*time.Hour),
				newComment(4, "mqube-bot", "Validation failed\n<!-- hash:  type: peacock-error -->\n", 2*time.Hour),
			},
		),
		ghmock.WithRequestMatchHandler(
			ghmock.DeleteReposIssuesCommentsByOwnerByRepoByCommentId,
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				deleted = append(deleted, r.URL.Path[strings.LastIndex(r.URL.Path, "/")+1:])
				w.WriteHeader(http.StatusNoContent)
			}),
		),
	)

	client := Client{github: github.NewClient(mockedHTTPClient), user: "mqube-bot"}
	err := client.DeletePeacockComments(context.Background(), "spring-financial-group", "peacock", 1)
	assert.NoError(t, err)
	assert.Equal(t, []string{"4", "3", "1"}, deleted)
}
