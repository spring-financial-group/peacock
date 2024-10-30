package webhookuc

import (
	"context"
	"github.com/google/go-github/v48/github"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spring-financial-group/peacock/pkg/config"
	"github.com/spring-financial-group/peacock/pkg/domain"
	"github.com/spring-financial-group/peacock/pkg/git/comment"
	"github.com/spring-financial-group/peacock/pkg/models"
	"strings"
)

type WebHookUseCase struct {
	cfg       *config.SCM
	scm       domain.SCM
	notesUC   domain.ReleaseNotesUseCase
	featherUC domain.FeathersUseCase
	releaseUC domain.ReleaseUseCase

	feathers map[int64]*feathersMeta
}

func NewUseCase(cfg *config.SCM, scm domain.SCM, notesUC domain.ReleaseNotesUseCase, feathersUC domain.FeathersUseCase, releaseUC domain.ReleaseUseCase) *WebHookUseCase {
	return &WebHookUseCase{
		cfg:       cfg,
		scm:       scm,
		notesUC:   notesUC,
		featherUC: feathersUC,
		feathers:  make(map[int64]*feathersMeta),
		releaseUC: releaseUC,
	}
}

func (w *WebHookUseCase) ValidatePeacock(e *models.PullRequestEventDTO) error {
	ctx := context.Background()

	// Set the current pipeline status to pending
	if err := w.createCommitStatus(ctx, e, domain.PendingState, e.SHA, domain.ValidationContext); err != nil {
		return w.handleError(ctx, domain.ValidationContext, e, errors.Wrap(err, "failed to create pending status"))
	}

	if e.Body == "" {
		log.Infof("no text found in PR body, skipping")
		return w.createCommitStatus(ctx, e, domain.SuccessState, e.SHA, domain.ValidationContext)
	}

	// Get the feathers for the pull request, should cache this as this will run for any edited event
	feathers, err := w.getFeathers(ctx, e.Branch, e)
	if err != nil {
		return w.handleError(ctx, domain.ValidationContext, e, err)
	}

	releaseNotes, err := w.notesUC.GetReleaseNotesFromMDAndTeams(e.Body, feathers.Teams)
	if err != nil {
		return w.handleError(ctx, domain.ValidationContext, e, errors.Wrap(err, "failed to parse release notes from markdown"))
	}
	if releaseNotes == nil {
		log.Infof("no releaseNotes found in PR body, skipping")
		return w.createCommitStatus(ctx, e, domain.SuccessState, e.SHA, domain.ValidationContext)
	}

	// Create a hash of the releaseNotes. Probably should cache these as well.
	newHash, err := w.notesUC.GenerateHash(releaseNotes)
	if err != nil {
		return w.handleError(ctx, domain.ValidationContext, e, errors.Wrap(err, "failed to generate message hash"))
	}

	// Get the hash from the last comment and compare
	comments, err := w.scm.GetPRCommentsByUser(ctx, e.RepoOwner, e.RepoName, e.PRNumber)
	if err != nil {
		return w.handleError(ctx, domain.ValidationContext, e, errors.Wrap(err, "failed to get comments"))
	}
	// Currently we only support one type of comment, so we can just get the most recent and check that
	if len(comments) > 0 {
		lastComment := comments[0]
		oldHash, _ := comment.GetMetadataFromComment(*lastComment.Body)
		if oldHash == newHash {
			log.Infof("message hash matches previous comment, skipping new breakdown")
			return w.createCommitStatus(ctx, e, domain.SuccessState, e.SHA, domain.ValidationContext)
		}
	}

	breakdown, err := w.notesUC.GenerateBreakdown(releaseNotes, newHash, len(feathers.Teams))
	if err != nil {
		return w.handleError(ctx, domain.ValidationContext, e, errors.Wrap(err, "failed to generate message breakdown"))
	}

	// We should prune the previous comments
	err = w.scm.DeleteUsersComments(ctx, e.RepoOwner, e.RepoName, e.PRNumber)
	if err != nil {
		return w.handleError(ctx, domain.ValidationContext, e, errors.Wrap(err, "failed to delete previous comments"))
	}

	// Comment on the PR with the breakdown
	log.Info("commenting on PR with message breakdown")
	err = w.scm.CommentOnPR(ctx, e.RepoOwner, e.RepoName, e.PRNumber, breakdown)
	if err != nil {
		return w.handleError(ctx, domain.ValidationContext, e, errors.Wrap(err, "failed to comment breakdown on PR"))
	}
	err = w.createCommitStatus(ctx, e, domain.SuccessState, e.SHA, domain.ValidationContext)
	if err != nil {
		log.Errorf("failed to create success status: %v", err)
	}
	return nil
}

func (w *WebHookUseCase) RunPeacock(e *models.PullRequestEventDTO) error {
	ctx := context.Background()
	defer w.CleanUp(e.PullRequestID)

	// We can use the most recent commit in the default branch to display the status. This way we don't have to worry about
	// merge method used on the PR. We'll continue to use the last commit SHA in the PR for error handling/feathers etc.
	defaultSHA, err := w.scm.GetLatestCommitSHAInBranch(ctx, e.RepoOwner, e.RepoName, e.DefaultBranch)
	if err != nil {
		return w.handleError(ctx, domain.ReleaseContext, e, errors.Wrap(err, "failed to get latest commit in default branch"))
	}

	// Set the current pipeline status to pending
	if err = w.createCommitStatus(ctx, e, domain.PendingState, defaultSHA, domain.ReleaseContext); err != nil {
		return w.handleError(ctx, domain.ReleaseContext, e, errors.Wrap(err, "failed to create pending status"))
	}

	if e.Body == "" {
		log.Infof("no text found in PR body, skipping")
		return w.createCommitStatus(ctx, e, domain.SuccessState, defaultSHA, domain.ReleaseContext)
	}

	// Get the feathers for the pull request, should cache this as this will run for any edited event
	feathers, err := w.getFeathers(ctx, e.DefaultBranch, e)
	if err != nil {
		return w.handleError(ctx, domain.ReleaseContext, e, err)
	}

	// Parse the PR body for any releaseNotes
	releaseNotes, err := w.notesUC.GetReleaseNotesFromMDAndTeams(e.Body, feathers.Teams)
	if err != nil {
		return w.handleError(ctx, domain.ReleaseContext, e, errors.Wrap(err, "failed to parse release notes from markdown"))
	}
	if releaseNotes == nil {
		log.Infof("no release notes found in PR body, skipping")
		return w.createCommitStatus(ctx, e, domain.SuccessState, defaultSHA, domain.ReleaseContext)
	}

	if err = w.notesUC.SendReleaseNotes(feathers.Config.Messages.Subject, releaseNotes); err != nil {
		return w.handleError(ctx, domain.ReleaseContext, e, errors.Wrap(err, "failed to send releaseNotes"))
	}

	log.Infof("%d message(s) sent", len(releaseNotes))

	files, err := w.scm.GetFilesChangedFromPR(ctx, e.RepoOwner, e.RepoName, e.PRNumber)
	if err != nil {
		return errors.Wrap(err, "failed to get changed files from pr")
	}

	if changedEnvironment := w.getChangedEnv(files); changedEnvironment != "" {
		log.Infof("saving release for environment %s", changedEnvironment)
		err = w.releaseUC.SaveRelease(ctx, changedEnvironment, releaseNotes, e.Summary())
		if err != nil {
			return w.handleError(ctx, domain.ReleaseContext, e, errors.Wrap(err, "failed to save release"))
		}
	} else {
		log.Warn("environment not found for release, skipping save")
	}

	err = w.createCommitStatus(ctx, e, domain.SuccessState, defaultSHA, domain.ReleaseContext)
	if err != nil {
		log.Errorf("failed to create success status: %v", err)
	}
	return nil
}

type feathersMeta struct {
	feathers *models.Feathers
	sha      string
}

func (w *WebHookUseCase) getFeathers(ctx context.Context, branch string, event *models.PullRequestEventDTO) (*models.Feathers, error) {
	// Get the feathers for the branch and check that it matches the sha
	meta, ok := w.feathers[event.PullRequestID]
	if ok && meta.sha == event.SHA {
		return meta.feathers, nil
	}

	meta = &feathersMeta{
		sha: event.SHA,
	}

	data, err := w.scm.GetFileFromBranch(ctx, event.RepoOwner, event.RepoName, branch, ".peacock/feathers.yaml")
	if err != nil {
		switch err.(type) {
		case *domain.ErrFileNotFound:
			return nil, errors.New("feathers does not exist in branch")
		default:
			return nil, err
		}
	}

	meta.feathers, err = w.featherUC.GetFeathersFromBytes(data)
	if err != nil {
		return nil, err
	}
	w.feathers[event.PullRequestID] = meta
	return meta.feathers, nil
}

func (w *WebHookUseCase) CleanUp(pullRequestID int64) {
	delete(w.feathers, pullRequestID)
}

func (w *WebHookUseCase) handleError(ctx context.Context, statusContext string, e *models.PullRequestEventDTO, err error) error {
	return w.scm.HandleError(ctx, statusContext, e.RepoOwner, e.RepoName, e.PRNumber, e.SHA, e.PROwner, err)
}

func (w *WebHookUseCase) createCommitStatus(ctx context.Context, e *models.PullRequestEventDTO, state domain.State, sha, context string) error {
	return w.scm.CreatePeacockCommitStatus(ctx, e.RepoOwner, e.RepoName, sha, state, context)
}

func (w *WebHookUseCase) getChangedEnv(files []*github.CommitFile) string {
	for _, file := range files {
		splitPath := strings.Split(*file.Filename, "/")
		for index, path := range splitPath {
			if path == "helmfiles" {
				return splitPath[index+1]
			}
		}
	}

	return ""
}
