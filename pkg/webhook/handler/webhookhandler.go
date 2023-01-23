package webhookhandler

import (
	"github.com/cbrgm/githubevents/githubevents"
	"github.com/gin-gonic/gin"
	"github.com/google/go-github/v48/github"
	log "github.com/sirupsen/logrus"
	"github.com/spring-financial-group/peacock/pkg/config"
	"github.com/spring-financial-group/peacock/pkg/models"
	"github.com/spring-financial-group/peacock/pkg/webhook/usecase"
)

type Handler struct {
	cfg     *config.SCM
	useCase *webhookuc.WebHookUseCase
	*githubevents.EventHandler
}

func NewHandler(cfg *config.SCM, group *gin.RouterGroup, uc *webhookuc.WebHookUseCase) {
	handler := &Handler{
		cfg:          cfg,
		useCase:      uc,
		EventHandler: githubevents.New(cfg.Secret),
	}
	handler.RegisterGitHubHooks()

	group.POST("/webhooks", handler.HandleEvents)
}

func (h *Handler) RegisterGitHubHooks() {
	h.OnPullRequestEventSynchronize(h.handlePullRequestSyncEvent)
	h.OnPullRequestEventOpened(h.handlePullRequestOpenedEvent)
	h.OnPullRequestEventReopened(h.handlePullRequestOpenedEvent)
	h.OnPullRequestEventClosed(h.handlePullRequestClosedEvent)
	h.OnPullRequestEventEdited(h.handlePullRequestEditEvent)
}

// HandleEvents godoc
// @Summary Endpoint for GitHub webhooks
// @Description Endpoint for GitHub webhooks
// @Tags webhook
// @Accept json
// @Success 200
// @Router /webhooks [post]
func (h *Handler) HandleEvents(c *gin.Context) {
	err := h.HandleEventRequest(c.Request)
	if err != nil {
		log.Error(err)
		return
	}
}

// handlePullRequestSyncEvent starts a dry-run when a PR has been synchronized (e.g. new commits pushed)
func (h *Handler) handlePullRequestSyncEvent(_ string, _ string, event *github.PullRequestEvent) error {
	log.Infof("%s/PR-%d synced. Starting dry-run.", *event.Repo.Name, *event.PullRequest.Number)
	return h.useCase.ValidatePeacock(models.MarshalPullRequestEvent(event))
}

// handlePullRequestOpenedEvent starts a dry-run when a PR has been opened
func (h *Handler) handlePullRequestOpenedEvent(_ string, _ string, event *github.PullRequestEvent) error {
	log.Infof("%s/PR-%d opened. Starting dry-run.", *event.Repo.Name, *event.PullRequest.Number)
	return h.useCase.ValidatePeacock(models.MarshalPullRequestEvent(event))
}

// handlePullRequestClosedEvent starts a full peacock run if the PR was merged, otherwise it removes the dangling
// feathers etc.
func (h *Handler) handlePullRequestClosedEvent(_ string, _ string, event *github.PullRequestEvent) error {
	if !*event.PullRequest.Merged {
		log.Infof("%s/PR-%d closed without merging. Skipping.", *event.Repo.Name, *event.PullRequest.Number)
		h.useCase.CleanUp(*event.PullRequest.ID)
		return nil
	}
	log.Infof("%s/PR-%d closed with merge. Starting full run.", *event.Repo.Name, *event.PullRequest.Number)
	return h.useCase.RunPeacock(models.MarshalPullRequestEvent(event))
}

// handlePullRequestEditEvent starts a dry-run when a PR has been edited (e.g. body/title changed)
func (h *Handler) handlePullRequestEditEvent(_ string, _ string, event *github.PullRequestEvent) error {
	log.Infof("%s/PR-%d edited. Starting dry-run.", *event.Repo.Name, *event.PullRequest.Number)
	return h.useCase.ValidatePeacock(models.MarshalPullRequestEvent(event))
}
