package webhookuc

import (
	"context"
	"fmt"
	"github.com/google/go-github/v48/github"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spring-financial-group/peacock/pkg/config"
	"github.com/spring-financial-group/peacock/pkg/domain"
	feather "github.com/spring-financial-group/peacock/pkg/feathers"
	"github.com/spring-financial-group/peacock/pkg/git/comment"
	"github.com/spring-financial-group/peacock/pkg/message"
	"strings"
)

type WebHookUseCase struct {
	cfg        *config.SCM
	scmFactory domain.SCMClientFactory
	handlers   map[string]domain.MessageHandler

	// Todo: cache feathers and messages better
	feathers map[string]*feathersInfo
}

func NewUseCase(cfg *config.SCM, scmFactory domain.SCMClientFactory, handlers map[string]domain.MessageHandler) *WebHookUseCase {
	return &WebHookUseCase{
		cfg:        cfg,
		scmFactory: scmFactory,
		handlers:   handlers,
		feathers:   make(map[string]*feathersInfo),
	}
}

func (w *WebHookUseCase) ValidatePeacock(event *github.PullRequestEvent) error {
	owner, repoName, prNumber, sha := *event.Repo.Owner.Login, *event.Repo.Name, *event.PullRequest.Number, *event.PullRequest.Head.SHA
	scm := w.scmFactory.GetClient(owner, repoName, w.cfg.User, prNumber)
	go w.createPendingStatus(context.Background(), scm, sha)

	ctx := context.Background()

	// Get the feathers for the pull request, should cache this as this will run for any edited event
	feathers, err := w.getFeathers(ctx, scm, *event.PullRequest.Head.Ref, sha)
	if err != nil {
		return w.handleError(ctx, scm, sha, err)
	}

	// Check that the relevant communication methods have been configured for the feathers
	types := feathers.GetAllContactTypes()
	if err != nil {
		return w.handleError(ctx, scm, sha, errors.Wrap(err, "failed to get contact types"))
	}
	for _, t := range types {
		_, ok := w.handlers[t]
		if !ok {
			return w.handleError(ctx, scm, sha, errors.New(fmt.Sprintf("message handler %s not found", t)))
		}
	}

	// Parse the PR body for any messages
	messages, err := message.ParseMessagesFromMarkdown(event.PullRequest.GetBody())
	if err != nil {
		return w.handleError(ctx, scm, sha, errors.Wrap(err, "failed to parse messages from markdown"))
	}
	if messages == nil {
		log.Infof("no messages found in PR body, skipping")
		return w.createSuccessStatus(ctx, scm, sha)
	}

	// Check that the teams in the messages exist in the feathers
	for _, m := range messages {
		if err = feathers.ExistsInFeathers(m.TeamNames...); err != nil {
			return w.handleError(ctx, scm, sha, errors.Wrap(err, "failed to find team in feathers"))
		}
	}

	// Create a hash of the messages. Probably should cache these as well.
	newHash, err := message.GenerateHash(messages)
	if err != nil {
		return w.handleError(ctx, scm, sha, errors.Wrap(err, "failed to generate message hash"))
	}

	// Get the hash from the last comment and compare
	comments, err := scm.GetPRCommentsByUser(ctx)
	if err != nil {
		return w.handleError(ctx, scm, sha, errors.Wrap(err, "failed to get comments"))
	}
	// Currently we only support one type of comment, so we can just get the most recent and check that
	if len(comments) > 0 {
		lastComment := comments[0]
		oldHash, _ := comment.GetMetadataFromComment(*lastComment.Body)
		if oldHash == newHash {
			log.Infof("message hash matches previous comment, skipping new breakdown")
			return nil
		}
	}

	breakdown, err := message.GenerateBreakdown(messages, len(feathers.Teams))
	if err != nil {
		return w.handleError(ctx, scm, sha, errors.Wrap(err, "failed to generate message breakdown"))
	}
	breakdown = comment.AddMetadataToComment(breakdown, newHash, comment.BreakdownCommentType)

	// We should prune the previous comments
	err = scm.DeleteUsersComments(ctx)
	if err != nil {
		return w.handleError(ctx, scm, sha, errors.Wrap(err, "failed to delete previous comments"))
	}

	// Comment on the PR with the breakdown
	err = scm.CommentOnPR(ctx, breakdown)
	if err != nil {
		return w.handleError(ctx, scm, sha, errors.Wrap(err, "failed to comment breakdown on PR"))
	}
	err = w.createSuccessStatus(ctx, scm, sha)
	if err != nil {
		log.Errorf("failed to create success status: %v", err)
	}
	return nil
}

func (w *WebHookUseCase) RunPeacock(event *github.PullRequestEvent) error {
	owner, repoName, prNumber, sha := *event.Repo.Owner.Login, *event.Repo.Name, *event.PullRequest.Number, *event.PullRequest.Head.SHA
	scm := w.scmFactory.GetClient(owner, repoName, w.cfg.User, prNumber)
	go w.createPendingStatus(context.Background(), scm, sha)

	ctx := context.Background()

	// Get the feathers for the pull request, should cache this as this will run for any edited event
	feathers, err := w.getFeathers(ctx, scm, *event.PullRequest.Head.Ref, sha)
	if err != nil {
		return w.handleError(ctx, scm, sha, err)
	}

	// Check that the relevant communication methods have been configured for the feathers
	types := feathers.GetAllContactTypes()
	if err != nil {
		return w.handleError(ctx, scm, sha, errors.Wrap(err, "failed to get contact types"))
	}
	for _, t := range types {
		_, ok := w.handlers[t]
		if !ok {
			return w.handleError(ctx, scm, sha, errors.New(fmt.Sprintf("message handler %s not found", t)))
		}
	}

	// Parse the PR body for any messages
	messages, err := message.ParseMessagesFromMarkdown(event.PullRequest.GetBody())
	if err != nil {
		return w.handleError(ctx, scm, sha, errors.Wrap(err, "failed to parse messages from markdown"))
	}
	if messages == nil {
		log.Infof("no messages found in PR body, skipping")
		return w.createSuccessStatus(ctx, scm, sha)
	}

	// Check that the teams in the messages exist in the feathers
	for _, m := range messages {
		if err = feathers.ExistsInFeathers(m.TeamNames...); err != nil {
			return w.handleError(ctx, scm, sha, errors.Wrap(err, "failed to find team in feathers"))
		}
	}

	if err = w.SendMessages(messages, feathers); err != nil {
		return w.handleError(ctx, scm, sha, errors.Wrap(err, "failed to send messages"))
	}

	err = w.createSuccessStatus(ctx, scm, sha)
	if err != nil {
		log.Errorf("failed to create success status: %v", err)
	}
	return nil
}

type feathersInfo struct {
	feathers *feather.Feathers
	sha      string
}

func (w *WebHookUseCase) getFeathers(ctx context.Context, scm domain.SCM, branch, sha string) (*feather.Feathers, error) {
	// Get the feathers for the branch and check that it matches the sha
	feathers, ok := w.feathers[branch]
	if ok && feathers.sha == sha {
		log.Infof("using cached feathers for branch %s", branch)
		return feathers.feathers, nil
	}

	log.Infof("fetching feathers for branch %s", branch)
	feathers = &feathersInfo{
		sha: sha,
	}

	data, err := scm.GetFileFromBranch(ctx, branch, ".peacock/feathers.yaml")
	if err != nil {
		return nil, err
	}
	feathers.feathers, err = feather.GetFeathersFromBytes(data)
	if err != nil {
		return nil, err
	}
	w.feathers[branch] = feathers
	return feathers.feathers, nil
}

func (w *WebHookUseCase) handleError(ctx context.Context, scm domain.SCM, headSHA string, err error) error {
	// With any errors we want to comment on the PR & set the status to failed
	commentErr := scm.CommentError(ctx, err)
	if commentErr != nil {
		log.Errorf("error commenting on PR: %v", commentErr)
		return err
	}

	status := &github.RepoStatus{
		State:       github.String("error"),
		Description: github.String(err.Error()),
		Context:     github.String("peacock-verify"),
	}
	statusErr := scm.CreateCommitStatus(ctx, headSHA, status)
	if statusErr != nil {
		log.Errorf("error setting commit status: %v", statusErr)
	}
	return err
}

func (w *WebHookUseCase) createSuccessStatus(ctx context.Context, scm domain.SCM, headSHA string) error {
	status := &github.RepoStatus{
		State:       github.String("success"),
		Description: github.String("Peacock verified"),
		Context:     github.String("peacock-verify"),
	}
	return scm.CreateCommitStatus(ctx, headSHA, status)
}

func (w *WebHookUseCase) createPendingStatus(ctx context.Context, scm domain.SCM, headSHA string) {
	status := &github.RepoStatus{
		State:       github.String("pending"),
		Description: github.String("Peacock verifying"),
		Context:     github.String("peacock-verify"),
	}
	err := scm.CreateCommitStatus(ctx, headSHA, status)
	if err != nil {
		log.Errorf("error setting pending commit status: %v", err)
	}
}

// SendMessages send the messages using the message handlers
func (w *WebHookUseCase) SendMessages(messages []message.Message, feathers *feather.Feathers) error {
	var errs []error
	for _, m := range messages {
		err := w.SendMessage(m, feathers)
		if err != nil {
			log.Error(err)
			errs = append(errs, err)
			continue
		}
	}
	if len(errs) > 0 {
		return errors.New("failed to send messages")
	}
	return nil
}

// SendMessage pools the addresses of the different teams by contactType and sends the message to each
func (w *WebHookUseCase) SendMessage(message message.Message, feathers *feather.Feathers) error {
	// We should pool the addresses by contact type so that we only send one message per contact type
	addressPool := feathers.GetAddressPoolByTeamNames(message.TeamNames...)
	for contactType, addresses := range addressPool {
		err := w.handlers[contactType].Send(message.Content, "test subject", addresses)
		if err != nil {
			return errors.Wrapf(err, "failed to send message")
		}
		log.Infof("Message successfully sent to %s via %s\n", strings.Join(addresses, ", "), contactType)
	}
	return nil
}
