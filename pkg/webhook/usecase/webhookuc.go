package webhookuc

import (
	"context"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spring-financial-group/peacock/pkg/config"
	"github.com/spring-financial-group/peacock/pkg/domain"
	feather "github.com/spring-financial-group/peacock/pkg/feathers"
	"github.com/spring-financial-group/peacock/pkg/git/comment"
	"github.com/spring-financial-group/peacock/pkg/models"
)

type WebHookUseCase struct {
	cfg        *config.SCM
	scmFactory domain.SCMClientFactory
	notesUC    domain.ReleaseNotesUseCase

	feathers map[int64]*feathersMeta
}

func NewUseCase(cfg *config.SCM, scmFactory domain.SCMClientFactory, notesUC domain.ReleaseNotesUseCase) *WebHookUseCase {
	return &WebHookUseCase{
		cfg:        cfg,
		scmFactory: scmFactory,
		notesUC:    notesUC,
		feathers:   make(map[int64]*feathersMeta),
	}
}

func (w *WebHookUseCase) ValidatePeacock(e *models.PullRequestEventDTO) error {
	ctx := context.Background()
	scm := w.scmFactory.GetClient(e.Owner, e.RepoName, w.cfg.User, e.PRNumber)
	defer w.scmFactory.RemoveClient(scm.GetKey())

	// Set the current pipeline status to pending
	if err := scm.CreatePeacockCommitStatus(ctx, e.SHA, domain.PendingState, domain.ValidationContext); err != nil {
		return scm.HandleError(ctx, domain.ValidationContext, e.SHA, errors.Wrap(err, "failed to create pending status"))
	}

	releaseNotes, err := w.notesUC.ParseNotesFromMarkdown(e.Body)
	if err != nil {
		return scm.HandleError(ctx, domain.ValidationContext, e.SHA, errors.Wrap(err, "failed to parse releaseNotes from markdown"))
	}
	if releaseNotes == nil {
		log.Infof("no releaseNotes found in PR body, skipping")
		return scm.CreatePeacockCommitStatus(ctx, e.SHA, domain.SuccessState, domain.ValidationContext)
	}

	// Get the feathers for the pull request, should cache this as this will run for any edited event
	feathers, err := w.getFeathers(ctx, scm, e.Branch, e)
	if err != nil {
		return scm.HandleError(ctx, domain.ValidationContext, e.SHA, err)
	}

	// Check that the relevant communication methods have been configured for the feathers
	if err = w.notesUC.ValidateReleaseNotesWithFeathers(feathers, releaseNotes); err != nil {
		return scm.HandleError(ctx, domain.ValidationContext, e.SHA, errors.Wrap(err, "failed to validate release notes with feathers"))
	}

	// Create a hash of the releaseNotes. Probably should cache these as well.
	newHash, err := w.notesUC.GenerateHash(releaseNotes)
	if err != nil {
		return scm.HandleError(ctx, domain.ValidationContext, e.SHA, errors.Wrap(err, "failed to generate message hash"))
	}

	// Get the hash from the last comment and compare
	comments, err := scm.GetPRCommentsByUser(ctx)
	if err != nil {
		return scm.HandleError(ctx, domain.ValidationContext, e.SHA, errors.Wrap(err, "failed to get comments"))
	}
	// Currently we only support one type of comment, so we can just get the most recent and check that
	if len(comments) > 0 {
		lastComment := comments[0]
		oldHash, _ := comment.GetMetadataFromComment(*lastComment.Body)
		if oldHash == newHash {
			log.Infof("message hash matches previous comment, skipping new breakdown")
			return scm.CreatePeacockCommitStatus(ctx, e.SHA, domain.SuccessState, domain.ValidationContext)
		}
	}

	breakdown, err := w.notesUC.GenerateBreakdown(releaseNotes, newHash, len(feathers.Teams))
	if err != nil {
		return scm.HandleError(ctx, domain.ValidationContext, e.SHA, errors.Wrap(err, "failed to generate message breakdown"))
	}

	// We should prune the previous comments
	err = scm.DeleteUsersComments(ctx)
	if err != nil {
		return scm.HandleError(ctx, domain.ValidationContext, e.SHA, errors.Wrap(err, "failed to delete previous comments"))
	}

	// Comment on the PR with the breakdown
	log.Info("commenting on PR with message breakdown")
	err = scm.CommentOnPR(ctx, breakdown)
	if err != nil {
		return scm.HandleError(ctx, domain.ValidationContext, e.SHA, errors.Wrap(err, "failed to comment breakdown on PR"))
	}
	err = scm.CreatePeacockCommitStatus(ctx, e.SHA, domain.SuccessState, domain.ValidationContext)
	if err != nil {
		log.Errorf("failed to create success status: %v", err)
	}
	return nil
}

func (w *WebHookUseCase) RunPeacock(e *models.PullRequestEventDTO) error {
	ctx := context.Background()
	scm := w.scmFactory.GetClient(e.Owner, e.RepoName, w.cfg.User, e.PRNumber)
	defer w.scmFactory.RemoveClient(scm.GetKey())
	defer w.CleanUp(e.PullRequestID)

	// We can use the most recent commit in the default branch to display the status. This way we don't have to worry about
	// merge method used on the PR. We'll continue to use the last commit SHA in the PR for error handling/feathers etc.
	defaultBranchSHA, err := scm.GetLatestCommitSHAInBranch(ctx, e.DefaultBranch)
	if err != nil {
		return scm.HandleError(ctx, domain.ReleaseContext, e.SHA, errors.Wrap(err, "failed to get latest commit in default branch"))
	}

	// Set the current pipeline status to pending
	if err = scm.CreatePeacockCommitStatus(ctx, defaultBranchSHA, domain.PendingState, domain.ReleaseContext); err != nil {
		return scm.HandleError(ctx, domain.ReleaseContext, e.SHA, errors.Wrap(err, "failed to create pending status"))
	}

	// Parse the PR body for any releaseNotes
	releaseNotes, err := w.notesUC.ParseNotesFromMarkdown(e.Body)
	if err != nil {
		return scm.HandleError(ctx, domain.ReleaseContext, e.SHA, errors.Wrap(err, "failed to parse releaseNotes from markdown"))
	}
	if releaseNotes == nil {
		log.Infof("no release notes found in PR body, skipping")
		return scm.CreatePeacockCommitStatus(ctx, defaultBranchSHA, domain.SuccessState, domain.ReleaseContext)
	}

	// Get the feathers for the pull request, should cache this as this will run for any edited event
	feathers, err := w.getFeathers(ctx, scm, e.DefaultBranch, e)
	if err != nil {
		return scm.HandleError(ctx, domain.ReleaseContext, e.SHA, err)
	}

	if err = w.notesUC.ValidateReleaseNotesWithFeathers(feathers, releaseNotes); err != nil {
		return scm.HandleError(ctx, domain.ReleaseContext, e.SHA, errors.Wrap(err, "failed to validate release notes with feathers"))
	}

	if err = w.notesUC.SendReleaseNotes(feathers, releaseNotes); err != nil {
		return scm.HandleError(ctx, domain.ReleaseContext, e.SHA, errors.Wrap(err, "failed to send releaseNotes"))
	}
	log.Infof("%d message(s) sent", len(releaseNotes))

	err = scm.CreatePeacockCommitStatus(ctx, defaultBranchSHA, domain.SuccessState, domain.ReleaseContext)
	if err != nil {
		log.Errorf("failed to create success status: %v", err)
	}
	return nil
}

type feathersMeta struct {
	feathers *feather.Feathers
	sha      string
}

func (w *WebHookUseCase) getFeathers(ctx context.Context, scm domain.SCM, branch string, event *models.PullRequestEventDTO) (*feather.Feathers, error) {
	// Get the feathers for the branch and check that it matches the sha
	meta, ok := w.feathers[event.PullRequestID]
	if ok && meta.sha == event.SHA {
		return meta.feathers, nil
	}

	meta = &feathersMeta{
		sha: event.SHA,
	}

	data, err := scm.GetFileFromBranch(ctx, branch, ".peacock/feathers.yaml")
	if err != nil {
		switch err.(type) {
		case *domain.ErrFileNotFound:
			return nil, errors.New("feathers does not exist in branch")
		default:
			return nil, err
		}
	}

	meta.feathers, err = feather.GetFeathersFromBytes(data)
	if err != nil {
		return nil, err
	}
	w.feathers[event.PullRequestID] = meta
	return meta.feathers, nil
}

func (w *WebHookUseCase) CleanUp(pullRequestID int64) {
	delete(w.feathers, pullRequestID)
}
