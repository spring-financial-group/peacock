package models

import (
	"github.com/google/go-github/v48/github"
)

// Pull request states
const (
	OpenState   = "open"
	ClosedState = "closed"
)

// IssueCommentEventDTO is a data transfer object for the IssueCommentEvent. It reduces the amount of data that is held
// by the service.
type IssueCommentEventDTO struct {
	PullRequestID int64
	PROwner       string
	RepoOwner     string
	RepoName      string
	Comment       string
	PRNumber      int
	SHA           string
	Branch        string
	DefaultBranch string
}

// PullRequestSummary is a summary of the PR details to be stored alongside release notes
type IssueCommentSummary struct {
	PRNumber  int
	RepoOwner string
	RepoName  string
}

func (p *IssueCommentEventDTO) Summary() IssueCommentSummary {
	return IssueCommentSummary{
		PRNumber:  p.PRNumber,
		RepoOwner: p.RepoOwner,
		RepoName:  p.RepoName,
	}
}

// MarshalIssueCommentCreatedEvent marshals a github.IssueCommentEvent into a IssueCommentEventDTO
func MarshalIssueCommentCreatedEvent(event *github.IssueCommentEvent) *IssueCommentEventDTO {
	return &IssueCommentEventDTO{
		PullRequestID: *event.Issue.ID,
		PROwner:       *event.Issue.User.Login,
		RepoOwner:     *event.Repo.Owner.Login,
		RepoName:      *event.Repo.Name,
		Comment:       *event.Comment.Body,
		PRNumber:      *event.Issue.Number,
		DefaultBranch: *event.Repo.DefaultBranch,
	}
}
