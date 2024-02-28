package models

import (
	"github.com/google/go-github/v48/github"
)

// Pull request states
const (
	OpenState   = "open"
	ClosedState = "closed"
)

// PullRequestEventDTO is a data transfer object for the PullRequestEvent. It reduces the amount of data that is held
// by the service.
type PullRequestEventDTO struct {
	PullRequestID int64
	PROwner       string
	RepoOwner     string
	RepoName      string
	Body          string
	PRNumber      int
	SHA           string
	Branch        string
	DefaultBranch string
}

// PullRequestSummary is a summary of the PR details to be stored alongside release notes
type PullRequestSummary struct {
	PRNumber  int
	RepoOwner string
	RepoName  string
}

func (p *PullRequestEventDTO) Summary() PullRequestSummary {
	return PullRequestSummary{
		PRNumber:  p.PRNumber,
		RepoOwner: p.RepoOwner,
		RepoName:  p.RepoName,
	}
}

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
