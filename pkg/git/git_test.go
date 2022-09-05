package git_test

import (
	"github.com/jenkins-x/jx-helpers/v3/pkg/gitclient/giturl"
	"github.com/spring-financial-group/peacock/pkg/git"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestName(t *testing.T) {
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
			_, err := git.NewClient(tt.serverURL, tt.owner, tt.repo, tt.token)
			if tt.shouldError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
