package models

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