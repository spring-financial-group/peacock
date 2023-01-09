package webhookuc

import (
	"context"
	"github.com/google/go-github/v48/github"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spring-financial-group/peacock/pkg/config"
	"github.com/spring-financial-group/peacock/pkg/domain"
	feather "github.com/spring-financial-group/peacock/pkg/feathers"
	"github.com/spring-financial-group/peacock/pkg/message"
)

type WebHookUseCase struct {
	cfg      *config.SCM
	git      domain.Git
	scm      domain.GitServer
	handlers map[string]domain.MessageHandler
}

func NewUseCase(cfg *config.SCM, git domain.Git, scm domain.GitServer, handlers map[string]domain.MessageHandler) *WebHookUseCase {
	return &WebHookUseCase{
		cfg:      cfg,
		git:      git,
		scm:      scm,
		handlers: handlers,
	}
}

func (w *WebHookUseCase) HandleDryRun(event *github.PullRequestEvent) error {
	ctx := context.Background()
	owner, repo := *event.Repo.Owner.Name, *event.Repo.Name

	// Get the feathers for the pull request, should cache this as this will run for any edited event
	feathers, err := w.getFeathers(ctx, owner, repo, *event.PullRequest.Head.Ref)
	if err != nil {
		return w.commentError(ctx, owner, repo, *event.PullRequest.Number, err)
	}

	// Check that the relevant communication methods have been configured for the feathers
	types := feathers.GetAllContactTypes()
	if err != nil {
		return w.commentError(ctx, owner, repo, *event.PullRequest.Number, errors.Wrap(err, "failed to get contact types"))
	}
	for _, t := range types {
		_, ok := w.handlers[t]
		if !ok {
			return w.commentError(ctx, owner, repo, *event.PullRequest.Number, errors.Wrapf(err, "message handler %s not found", t))
		}
	}

	// Parse the PR body for any messages
	messages, err := message.ParseMessagesFromMarkdown(*event.PullRequest.Body)
	if err != nil {
		return w.commentError(ctx, owner, repo, *event.PullRequest.Number, errors.Wrap(err, "failed to parse messages from markdown"))
	}
	if messages == nil {
		log.Infof("no messages found in PR body")
		return nil
	}

	// Check that the teams in the messages exist in the feathers
	for _, m := range messages {
		if err = feathers.ExistsInFeathers(m.TeamNames...); err != nil {
			return w.commentError(ctx, owner, repo, *event.PullRequest.Number, errors.Wrap(err, "failed to find team in feathers"))
		}
	}

	// Create a hash of the messages. Probably should cache these as well.
	hash, err := message.GenerateHash(messages)
	if err != nil {
		return w.commentError(ctx, owner, repo, *event.PullRequest.Number, errors.Wrap(err, "failed to generate message hash"))
	}

	breakdown, err := message.GenerateBreakdown(messages, len(feathers.Teams), hash)
	if err != nil {
		return w.commentError(ctx, owner, repo, *event.PullRequest.Number, errors.Wrap(err, "failed to generate message breakdown"))
	}

	// Comment on the PR with the breakdown
	err = w.scm.CommentOnPR(ctx, owner, repo, *event.PullRequest.Number, breakdown)
	if err != nil {
		return w.commentError(ctx, owner, repo, *event.PullRequest.Number, errors.Wrap(err, "failed to comment breakdown on PR"))
	}
	return nil
}

func (w *WebHookUseCase) HandlePRMerge(event *github.PullRequestEvent) error {
	// Should be pretty simple. If we cache all the info about the PR, we can just send the messages.
	return nil
}

func (w *WebHookUseCase) getFeathers(ctx context.Context, owner, repo, branch string) (*feather.Feathers, error) {
	// Get the feathers for the pull request, should cache this as this will run for any edited event
	data, err := w.scm.GetFileFromBranch(ctx, owner, repo, branch, ".peacock/feathers.yaml")
	if err != nil {
		return nil, err
	}
	return feather.GetFeathersFromBytes(data)
}

func (w *WebHookUseCase) commentError(ctx context.Context, owner, repo string, prNumber int, err error) error {
	commentErr := w.scm.CommentError(ctx, owner, repo, prNumber, err)
	if commentErr != nil {
		log.Errorf("error commenting on PR: %v", commentErr)
	}
	return err
}
