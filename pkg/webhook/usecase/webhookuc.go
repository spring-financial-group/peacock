package webhookuc

import (
	"context"
	"strings"

	"github.com/google/go-github/v48/github"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spring-financial-group/peacock/pkg/config"
	"github.com/spring-financial-group/peacock/pkg/domain"
	"github.com/spring-financial-group/peacock/pkg/git/comment"
	"github.com/spring-financial-group/peacock/pkg/models"
)

type WebHookUseCase struct {
	cfg       *config.SCM
	scm       domain.SCM
	notesUC   domain.ReleaseNotesUseCase
	featherUC domain.FeathersUseCase
	releaseUC domain.ReleaseUseCase

	feathers    map[int64]*feathersMeta
	prTemplates map[int64]*prTemplateMeta
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

	// Get PR for SHA
	if e.Branch == "" {
		branch, sha, err := w.scm.GetPRBranchSHAFromPRNumber(ctx, e.RepoOwner, e.RepoName, e.PRNumber)
		if err != nil {
			println("failed to get PR details")
			return w.handleError(ctx, domain.ValidationContext, e, err)
		}
		e.Branch = *branch
		e.SHA = *sha
	}

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

	releaseNotes, err := w.notesUC.GetReleaseNotesFromMarkdownAndTeamsInFeathers(e.Body, feathers.Teams)
	if err != nil {
		return w.handleError(ctx, domain.ValidationContext, e, errors.Wrap(err, "failed to parse release notes from markdown"))
	}
	if releaseNotes == nil {
		log.Infof("no releaseNotes found in PR body, skipping")
		return w.createCommitStatus(ctx, e, domain.SuccessState, e.SHA, domain.ValidationContext)
	}

	// ensure the current notes aren't the same as the template, as we don't want default messages being sent out
	templateReleaseNotes, err := w.getPRTemplateOrDefault(ctx, e.Branch, e, feathers.Teams)
	if err != nil {
		return w.handleError(ctx, domain.ValidationContext, e, err)
	}

	areSame, err := w.compareReleaseNotesAndTemplate(releaseNotes, templateReleaseNotes)
	if err != nil {
		return w.handleError(ctx, domain.ValidationContext, e, err)
	}

	if areSame {
		log.Infof("release notes are same as the pull request template, failing")
		return w.handleError(ctx, domain.ValidationContext, e, errors.New("release notes cannot be the same as the pull request template"))
	}

	// Prevent doing work if the new release notes are same as the previous release notes
	newHash, err := w.notesUC.GenerateHash(releaseNotes)
	if err != nil {
		return w.handleError(ctx, domain.ValidationContext, e, errors.Wrap(err, "failed to generate message hash"))
	}

	// Compare the previous hash to the current one, stop here if there are no changes (there's no work to do)
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

	// Break down the release notes to prove we've parsed them and to check the formatting
	breakdown, err := w.notesUC.GenerateBreakdown(releaseNotes, newHash, len(feathers.Teams))
	if err != nil {
		return w.handleError(ctx, domain.ValidationContext, e, errors.Wrap(err, "failed to generate message breakdown"))
	}

	// prune the previous comments to prevent spam
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
	releaseNotes, err := w.notesUC.GetReleaseNotesFromMarkdownAndTeamsInFeathers(e.Body, feathers.Teams)
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
		var errFileNotFound *domain.ErrFileNotFound
		switch {
		case errors.As(err, &errFileNotFound):
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

type prTemplateMeta struct {
	prTemplates []models.ReleaseNote
	sha         string
}

func (w *WebHookUseCase) getPRTemplateOrDefault(ctx context.Context, branch string, event *models.PullRequestEventDTO, teamsInFeathers models.Teams) ([]models.ReleaseNote, error) {
	// Get the release notes for the branch and check that it matches the sha
	meta, ok := w.prTemplates[event.PullRequestID]
	if ok && meta.sha == event.SHA {
		return meta.prTemplates, nil
	}

	meta = &prTemplateMeta{
		sha: event.SHA,
	}

	data, err := w.scm.GetFileFromBranch(ctx, event.RepoOwner, event.RepoName, branch, ".github/pull_request_template.md")
	if err != nil {
		var errFileNotFound *domain.ErrFileNotFound
		switch {
		// is this actually an error as you could peacock without a template
		case errors.As(err, &errFileNotFound):
			log.Infof("PR template not found, continuing with default")
			return []models.ReleaseNote{}, nil
		default:
			return nil, err
		}
	}

	meta.prTemplates, err = w.notesUC.GetReleaseNotesFromMarkdownAndTeamsInFeathers(string(data[:]), teamsInFeathers)
	if err != nil {
		return nil, err
	}
	w.prTemplates[event.PullRequestID] = meta
	return meta.prTemplates, nil
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

func (w *WebHookUseCase) compareReleaseNotesAndTemplate(
	notes []models.ReleaseNote,
	templates []models.ReleaseNote,
) (bool, error) {

	if len(notes) != len(templates) {
		return false, nil
	}

	// Index templates by content (or another stable key if you have one)
	templateIndex := make(map[string]models.ReleaseNote, len(templates))
	for _, t := range templates {
		templateIndex[t.Content] = t
	}

	for _, note := range notes {
		tmpl, ok := templateIndex[note.Content]
		if !ok {
			return false, nil
		}

		if !teamsEqual(note.Teams, tmpl.Teams) {
			return false, nil
		}
	}

	return true, nil
}

func teamsEqual(a, b models.Teams) bool {
	if len(a) != len(b) {
		return false
	}

	index := make(map[string]models.Team, len(a))
	for _, team := range a {
		index[team.Name] = team
	}

	for _, team := range b {
		existing, ok := index[team.Name]
		if !ok {
			return false
		}

		if !teamEqual(existing, team) {
			return false
		}
	}

	return true
}

func teamEqual(a, b models.Team) bool {
	if a.Name != b.Name ||
		a.APIKey != b.APIKey ||
		a.ContactType != b.ContactType {
		return false
	}

	return stringSliceEqual(a.Addresses, b.Addresses)
}

func stringSliceEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	m := make(map[string]int, len(a))
	for _, v := range a {
		m[v]++
	}

	for _, v := range b {
		if m[v] == 0 {
			return false
		}
		m[v]--
	}

	return true
}
