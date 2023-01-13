package models

import (
	"github.com/google/go-github/v48/github"
)

// PullRequestEventDTO is a data transfer object for the PullRequestEvent. It reduces the amount of data that is held
// by the service.
type PullRequestEventDTO struct {
	Owner    string
	RepoName string
	Body     string
	PRNumber int
	SHA      string
	Branch   string
}

// MarshalPullRequestEvent marshals a github.PullRequestEvent into a PullRequestEventDTO
func MarshalPullRequestEvent(event *github.PullRequestEvent) *PullRequestEventDTO {
	return &PullRequestEventDTO{
		Owner:    *event.Repo.Owner.Login,
		RepoName: *event.Repo.Name,
		Body:     *event.PullRequest.Body,
		PRNumber: *event.PullRequest.Number,
		SHA:      *event.PullRequest.Head.SHA,
		Branch:   *event.PullRequest.Head.Ref,
	}
}