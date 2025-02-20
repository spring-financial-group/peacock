package models

import (
	"github.com/google/go-github/v48/github"
)

// MarshalIssueCommentCreatedEvent marshals a github.IssueCommentEvent into a PullRequestEventDTO
func MarshalIssueCommentCreatedEvent(event *github.IssueCommentEvent) *PullRequestEventDTO {
	return &PullRequestEventDTO{
		PullRequestID: *event.Issue.ID,
		PROwner:       *event.Issue.User.Login,
		RepoOwner:     *event.Repo.Owner.Login,
		RepoName:      *event.Repo.Name,
		Body:          *event.Comment.Body,
		PRNumber:      *event.Issue.Number,
		DefaultBranch: *event.Repo.DefaultBranch,
	}
}
