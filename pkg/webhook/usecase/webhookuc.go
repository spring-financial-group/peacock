package webhookuc

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spring-financial-group/peacock/pkg/config"
	"github.com/spring-financial-group/peacock/pkg/domain"
	feather "github.com/spring-financial-group/peacock/pkg/feathers"
	"github.com/spring-financial-group/peacock/pkg/git/comment"
	"github.com/spring-financial-group/peacock/pkg/message"
	"github.com/spring-financial-group/peacock/pkg/models"
	"strings"
)

type WebHookUseCase struct {
	cfg        *config.SCM
	scmFactory domain.SCMClientFactory
	handlers   map[string]domain.MessageHandler

	feathers map[string]*feathersMeta
}

func NewUseCase(cfg *config.SCM, scmFactory domain.SCMClientFactory, handlers map[string]domain.MessageHandler) *WebHookUseCase {
	return &WebHookUseCase{
		cfg:        cfg,
		scmFactory: scmFactory,
		handlers:   handlers,
		feathers:   make(map[string]*feathersMeta),
	}
}

func (w *WebHookUseCase) ValidatePeacock(e *models.PullRequestEventDTO) error {
	ctx := context.Background()
	scm := w.scmFactory.GetClient(e.Owner, e.RepoName, w.cfg.User, e.PRNumber)
	defer w.scmFactory.RemoveClient(scm.GetKey())

	// Set the current pipeline status to pending
	if err := scm.CreatePeacockCommitStatus(ctx, e.SHA, domain.PendingStatus, domain.ValidationContext); err != nil {
		return scm.HandleError(ctx, domain.ValidationContext, e.SHA, errors.Wrap(err, "failed to create pending status"))
	}

	// Get the feathers for the pull request, should cache this as this will run for any edited event
	feathers, err := w.getFeathers(ctx, scm, e.Branch, e.SHA)
	if err != nil {
		return scm.HandleError(ctx, domain.ValidationContext, e.SHA, err)
	}

	// Check that the relevant communication methods have been configured for the feathers
	types := feathers.GetAllContactTypes()
	if err != nil {
		return scm.HandleError(ctx, domain.ValidationContext, e.SHA, errors.Wrap(err, "failed to get contact types"))
	}
	for _, t := range types {
		_, ok := w.handlers[t]
		if !ok {
			return scm.HandleError(ctx, domain.ValidationContext, e.SHA, errors.New(fmt.Sprintf("message handler %s not found", t)))
		}
	}

	messages, err := message.ParseMessagesFromMarkdown(e.Body)
	if err != nil {
		return scm.HandleError(ctx, domain.ValidationContext, e.SHA, errors.Wrap(err, "failed to parse messages from markdown"))
	}
	if messages == nil {
		log.Infof("no messages found in PR body, skipping")
		return scm.CreatePeacockCommitStatus(ctx, e.SHA, domain.SuccessStatus, domain.ValidationContext)
	}

	// Check that the teams in the messages exist in the feathers
	for _, m := range messages {
		if err = feathers.ExistsInFeathers(m.TeamNames...); err != nil {
			return scm.HandleError(ctx, domain.ValidationContext, e.SHA, errors.Wrap(err, "failed to find team in feathers"))
		}
	}

	// Create a hash of the messages. Probably should cache these as well.
	newHash, err := message.GenerateHash(messages)
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
			return scm.CreatePeacockCommitStatus(ctx, e.SHA, domain.SuccessStatus, domain.ValidationContext)
		}
	}

	breakdown, err := message.GenerateBreakdown(messages, len(feathers.Teams))
	if err != nil {
		return scm.HandleError(ctx, domain.ValidationContext, e.SHA, errors.Wrap(err, "failed to generate message breakdown"))
	}
	breakdown = comment.AddMetadataToComment(breakdown, newHash, comment.BreakdownCommentType)

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
	err = scm.CreatePeacockCommitStatus(ctx, e.SHA, domain.SuccessStatus, domain.ValidationContext)
	if err != nil {
		log.Errorf("failed to create success status: %v", err)
	}
	return nil
}

func (w *WebHookUseCase) RunPeacock(e *models.PullRequestEventDTO) error {
	ctx := context.Background()
	scm := w.scmFactory.GetClient(e.Owner, e.RepoName, w.cfg.User, e.PRNumber)
	defer w.scmFactory.RemoveClient(scm.GetKey())

	// We can use the most recent commit in the default branch to display the status. This way we don't have to worry about
	// merge method used on the PR. We'll continue to use the last commit SHA in the PR for error handling/feathers etc.
	latestCommit, err := scm.GetLatestCommitInBranch(ctx, e.DefaultBranch)
	if err != nil {
		return scm.HandleError(ctx, domain.ReleaseContext, e.SHA, errors.Wrap(err, "failed to get latest commit in default branch"))
	}
	defaultBranchSHA := *latestCommit.SHA

	// Set the current pipeline status to pending
	if err = scm.CreatePeacockCommitStatus(ctx, defaultBranchSHA, domain.PendingStatus, domain.ReleaseContext); err != nil {
		return scm.HandleError(ctx, domain.ReleaseContext, e.SHA, errors.Wrap(err, "failed to create pending status"))
	}

	// Get the feathers for the pull request, should cache this as this will run for any edited event
	feathers, err := w.getFeathers(ctx, scm, e.Branch, e.SHA)
	if err != nil {
		return scm.HandleError(ctx, domain.ReleaseContext, domain.ValidationContext, err)
	}

	// Check that the relevant communication methods have been configured for the feathers
	types := feathers.GetAllContactTypes()
	if err != nil {
		return scm.HandleError(ctx, domain.ReleaseContext, e.SHA, errors.Wrap(err, "failed to get contact types"))
	}
	for _, t := range types {
		_, ok := w.handlers[t]
		if !ok {
			return scm.HandleError(ctx, domain.ReleaseContext, e.SHA, errors.New(fmt.Sprintf("message handler %s not found", t)))
		}
	}

	// Parse the PR body for any messages
	messages, err := message.ParseMessagesFromMarkdown(e.Body)
	if err != nil {
		return scm.HandleError(ctx, domain.ReleaseContext, e.SHA, errors.Wrap(err, "failed to parse messages from markdown"))
	}
	if messages == nil {
		log.Infof("no messages found in PR body, skipping")
		return scm.CreatePeacockCommitStatus(ctx, defaultBranchSHA, domain.SuccessStatus, domain.ReleaseContext)
	}

	// Check that the teams in the messages exist in the feathers
	for _, m := range messages {
		if err = feathers.ExistsInFeathers(m.TeamNames...); err != nil {
			return scm.HandleError(ctx, domain.ReleaseContext, e.SHA, errors.Wrap(err, "failed to find team in feathers"))
		}
	}

	if err = w.SendMessages(messages, feathers); err != nil {
		return scm.HandleError(ctx, domain.ReleaseContext, e.SHA, errors.Wrap(err, "failed to send messages"))
	}
	log.Infof("%d message(s) sent", len(messages))

	// Once the messages have been sent we can remove the cached feathers
	w.CleanUp(e.Branch)

	err = scm.CreatePeacockCommitStatus(ctx, defaultBranchSHA, domain.SuccessStatus, domain.ReleaseContext)
	if err != nil {
		log.Errorf("failed to create success status: %v", err)
	}
	return nil
}

type feathersMeta struct {
	feathers *feather.Feathers
	sha      string
}

func (w *WebHookUseCase) getFeathers(ctx context.Context, scm domain.SCM, branch, sha string) (*feather.Feathers, error) {
	// Get the feathers for the branch and check that it matches the sha
	meta, ok := w.feathers[branch]
	if ok && meta.sha == sha {
		return meta.feathers, nil
	}

	meta = &feathersMeta{
		sha: sha,
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
	w.feathers[branch] = meta
	return meta.feathers, nil
}

func (w *WebHookUseCase) CleanUp(branch string) {
	delete(w.feathers, branch)
}

// SendMessages send the messages using the message handlers
func (w *WebHookUseCase) SendMessages(messages []message.Message, feathers *feather.Feathers) error {
	var errCount int
	for _, m := range messages {
		err := w.sendMessage(m, feathers)
		if err != nil {
			log.Error(err)
			errCount++
			continue
		}
	}
	if errCount > 0 {
		return errors.New("failed to send messages")
	}
	return nil
}

// sendMessage pools the addresses of the different teams by contactType and sends the message to each
func (w *WebHookUseCase) sendMessage(message message.Message, feathers *feather.Feathers) error {
	// We should pool the addresses by contact type so that we only send one message per contact type
	addressPool := feathers.GetAddressPoolByTeamNames(message.TeamNames...)
	for contactType, addresses := range addressPool {
		err := w.handlers[contactType].Send(message.Content, feathers.Config.Messages.Subject, addresses)
		if err != nil {
			return errors.Wrapf(err, "failed to send message")
		}
		log.Infof("Message successfully sent to %s via %s", strings.Join(addresses, ", "), contactType)
	}
	return nil
}
