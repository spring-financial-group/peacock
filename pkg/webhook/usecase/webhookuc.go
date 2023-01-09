package webhookuc

import (
	"context"
	"github.com/google/go-github/v48/github"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spring-financial-group/peacock/pkg/config"
	"github.com/spring-financial-group/peacock/pkg/domain"
	feather "github.com/spring-financial-group/peacock/pkg/feathers"
	"github.com/spring-financial-group/peacock/pkg/git/comment"
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
	owner, repoName, prNumber := *event.Repo.Owner.Login, *event.Repo.Name, *event.PullRequest.Number

	// Get the feathers for the pull request, should cache this as this will run for any edited event
	feathers, err := w.getFeathers(ctx, owner, repoName, *event.PullRequest.Head.Ref)
	if err != nil {
		return w.commentError(ctx, owner, repoName, prNumber, err)
	}

	// Check that the relevant communication methods have been configured for the feathers
	types := feathers.GetAllContactTypes()
	if err != nil {
		return w.commentError(ctx, owner, repoName, prNumber, errors.Wrap(err, "failed to get contact types"))
	}
	for _, t := range types {
		_, ok := w.handlers[t]
		if !ok {
			return w.commentError(ctx, owner, repoName, prNumber, errors.Wrapf(err, "message handler %s not found", t))
		}
	}

	// Parse the PR body for any messages
	messages, err := message.ParseMessagesFromMarkdown(event.PullRequest.GetBody())
	if err != nil {
		return w.commentError(ctx, owner, repoName, prNumber, errors.Wrap(err, "failed to parse messages from markdown"))
	}
	if messages == nil {
		log.Infof("no messages found in PR body")
		return nil
	}

	// Check that the teams in the messages exist in the feathers
	for _, m := range messages {
		if err = feathers.ExistsInFeathers(m.TeamNames...); err != nil {
			return w.commentError(ctx, owner, repoName, prNumber, errors.Wrap(err, "failed to find team in feathers"))
		}
	}

	// Create a hash of the messages. Probably should cache these as well.
	newHash, err := message.GenerateHash(messages)
	if err != nil {
		return w.commentError(ctx, owner, repoName, prNumber, errors.Wrap(err, "failed to generate message hash"))
	}

	// Get the hash from the last comment and compare
	comments, err := w.scm.GetPRCommentsByUser(ctx, owner, repoName, w.cfg.User, prNumber)
	if err != nil {
		return w.commentError(ctx, owner, repoName, prNumber, errors.Wrap(err, "failed to get comments"))
	}
	// Currently we only support one type of comment, so we can just get the most recent and check that
	if len(comments) > 0 {
		lastComment := comments[0]
		oldHash, _ := comment.GetMetadataFromComment(*lastComment.Body)
		if oldHash == newHash {
			log.Infof("hashes match, no need to comment")
			return nil
		}
	}

	breakdown, err := message.GenerateBreakdown(messages, len(feathers.Teams))
	if err != nil {
		return w.commentError(ctx, owner, repoName, prNumber, errors.Wrap(err, "failed to generate message breakdown"))
	}
	breakdown = comment.AddMetadataToComment(breakdown, newHash, comment.BreakdownCommentType)

	// We should prune the previous comments
	err = w.scm.DeleteUsersComments(ctx, owner, repoName, w.cfg.User, prNumber)
	if err != nil {
		return w.commentError(ctx, owner, repoName, prNumber, errors.Wrap(err, "failed to delete previous comments"))
	}

	// Comment on the PR with the breakdown
	err = w.scm.CommentOnPR(ctx, owner, repoName, prNumber, breakdown)
	if err != nil {
		return w.commentError(ctx, owner, repoName, prNumber, errors.Wrap(err, "failed to comment breakdown on PR"))
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
