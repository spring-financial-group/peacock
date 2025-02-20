package models

import (
	"github.com/google/go-github/v48/github"
)

// MarshalPullRequestEvent marshals a github.PullRequestEvent into a PullRequestEventDTO
func MarshalPullRequestEvent(event *github.PullRequestEvent) *PullRequestEventDTO {
	return &PullRequestEventDTO{
		PullRequestID: *event.PullRequest.ID,
		PROwner:       *event.PullRequest.User.Login,
		RepoOwner:     *event.Repo.Owner.Login,
		RepoName:      *event.Repo.Name,
		Body:          event.PullRequest.GetBody(),
		PRNumber:      *event.PullRequest.Number,
		SHA:           *event.PullRequest.Head.SHA,
		Branch:        *event.PullRequest.Head.Ref,
		DefaultBranch: *event.Repo.DefaultBranch,
	}
}
