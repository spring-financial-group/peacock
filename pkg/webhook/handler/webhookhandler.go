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
	h.OnPullRequestEventAny(h.logPullRequestEvent)
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

func (h *Handler) logPullRequestEvent(_ string, eventName string, event *github.PullRequestEvent) error {
	log.Infof("Received %s event for %s-PR%d", eventName, *event.Repo.Name, *event.PullRequest.Number)
	return nil
}

func (h *Handler) handlePullRequestOpenedEvent(_ string, _ string, event *github.PullRequestEvent) error {
	log.Info("PR opened. Starting dry-run.")
	return h.useCase.ValidatePeacock(models.MarshalPullRequestEvent(event))
}

func (h *Handler) handlePullRequestClosedEvent(_ string, _ string, event *github.PullRequestEvent) error {
	if !*event.PullRequest.Merged {
		log.Info("PR closed without merging. Skipping.")
		h.useCase.CleanUp(*event.PullRequest.Head.Ref)
		return nil
	}
	log.Info("PR merged. Starting full run.")
	return h.useCase.RunPeacock(models.MarshalPullRequestEvent(event))
}

func (h *Handler) handlePullRequestEditEvent(_ string, _ string, event *github.PullRequestEvent) error {
	log.Info("PR edited. Starting dry-run.")
	return h.useCase.ValidatePeacock(models.MarshalPullRequestEvent(event))
}
