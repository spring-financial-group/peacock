package webhook

import (
	"github.com/cbrgm/githubevents/githubevents"
	"github.com/gin-gonic/gin"
	"github.com/google/go-github/v48/github"
	log "github.com/sirupsen/logrus"
	"github.com/spring-financial-group/peacock/pkg/config"
)

const (
	editedAction = "edited"
)

type Handler struct {
	cfg *config.SCM
	*githubevents.EventHandler
}

func NewHandler(cfg *config.SCM) (*Handler, error) {
	handler := &Handler{
		cfg:          cfg,
		EventHandler: githubevents.New(cfg.Secret),
	}
	return handler, handler.RegisterHooks()
}

func (h *Handler) RegisterHooks() error {
	h.OnPullRequestEventAny(
		func(deliveryID string, eventName string, event *github.PullRequestEvent) error {
			// Edited events aren't supported by the githubevents package, so we need to manually filter for them
			if *event.Action == editedAction {
				// Todo: add edited function
				log.Info("Received pull request edited action")
			} else {
				log.Info("Received pull request action: ", *event.Action)
			}
			return nil
		},
	)
	return nil
}

func (h *Handler) HandleEvents(c *gin.Context) {
	err := h.HandleEventRequest(c.Request)
	if err != nil {
		log.Error(err)
		return
	}
}
