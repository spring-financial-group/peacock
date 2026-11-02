package github

import (
	"context"
	"encoding/base64"
	"fmt"
	"github.com/google/go-github/v48/github"
	ghmock "github.com/migueleliasweb/go-github-mock/src/mock"
	"github.com/pkg/errors"
	"github.com/spring-financial-group/peacock/pkg/domain"
	"github.com/spring-financial-group/peacock/pkg/utils"
	"github.com/stretchr/testify/assert"
	"net/http"
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

func TestGit_GetPRBaseBranchFromPRNumber(t *testing.T) {
	mockedHTTPClient := ghmock.NewMockedHTTPClient(
		ghmock.WithRequestMatch(
			ghmock.GetReposPullsByOwnerByRepoByPullNumber,
			&github.PullRequest{
				Base: &github.PullRequestBranch{
					Ref: utils.NewPtr("master"),
				},
			},
		),
	)
	client := Client{
		github: github.NewClient(mockedHTTPClient),
		user:   "mqube-bot",
	}

	actual, err := client.GetPRBaseBranchFromPRNumber(context.Background(), "spring-financial-group", "peacock", 1)
	assert.NoError(t, err)
	assert.Equal(t, utils.NewPtr("master"), actual)
}

func TestGit_GetFileFromBranch(t *testing.T) {
	baseContent := base64.StdEncoding.EncodeToString([]byte(`{"base": true}`))
	headContent := base64.StdEncoding.EncodeToString([]byte(`{"head": true}`))

	mockedHTTPClient := ghmock.NewMockedHTTPClient(
		ghmock.WithRequestMatchHandler(
			ghmock.GetReposContentsByOwnerByRepoByPath,
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch ref := r.URL.Query().Get("ref"); ref {
				case "master":
					_, _ = fmt.Fprintf(w, `{"name":"schema.json","path":"schema.json","content":%q,"encoding":"base64"}`, baseContent)
				case "missing":
					w.WriteHeader(http.StatusNotFound)
				default:
					_, _ = fmt.Fprintf(w, `{"name":"schema.json","path":"schema.json","content":%q,"encoding":"base64"}`, headContent)
				}
			}),
		),
	)
	client := Client{
		github: github.NewClient(mockedHTTPClient),
		user:   "mqube-bot",
	}

	t.Run("file is fetched from the base ref", func(t *testing.T) {
		actual, err := client.GetFileFromBranch(context.Background(), "spring-financial-group", "peacock", "master", "schema.json")
		assert.NoError(t, err)
		assert.Equal(t, `{"base": true}`, string(actual))
	})

	t.Run("file is fetched from the head sha", func(t *testing.T) {
		actual, err := client.GetFileFromBranch(context.Background(), "spring-financial-group", "peacock", "head-sha", "schema.json")
		assert.NoError(t, err)
		assert.Equal(t, `{"head": true}`, string(actual))
	})

	t.Run("missing files map to ErrFileNotFound", func(t *testing.T) {
		_, err := client.GetFileFromBranch(context.Background(), "spring-financial-group", "peacock", "missing", "schema.json")
		assert.Error(t, err)
		var notFound *domain.ErrFileNotFound
		assert.True(t, errors.As(err, &notFound))
	})
}
