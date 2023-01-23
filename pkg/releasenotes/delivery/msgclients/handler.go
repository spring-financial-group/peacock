package msgclients

import (
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spring-financial-group/peacock/pkg/config"
	"github.com/spring-financial-group/peacock/pkg/domain"
	feather "github.com/spring-financial-group/peacock/pkg/feathers"
	"github.com/spring-financial-group/peacock/pkg/models"
	"github.com/spring-financial-group/peacock/pkg/msgclients/slack"
	"github.com/spring-financial-group/peacock/pkg/msgclients/webhook"
	"strings"
)

type Handler struct {
	Clients map[string]domain.MessageClient
}

func NewMessageHandler(cfg *config.MessageHandlers) *Handler {
	clients := make(map[string]domain.MessageClient)
	if cfg.Slack.Token != "" {
		log.Info("Slack message handler initialised")
		clients[models.Slack] = slack.NewClient(cfg.Slack.Token)
	}
	if cfg.Webhook.URL != "" && cfg.Webhook.Secret != "" {
		log.Info("Webhook message handler initialised")
		clients[models.Webhook] = webhook.NewClient(cfg.Webhook.URL, cfg.Webhook.Token, cfg.Webhook.Secret)
	}
	return &Handler{
		Clients: clients,
	}
}

func (h *Handler) SendReleaseNotes(feathers *feather.Feathers, notes []models.ReleaseNote) error {
	var errCount int
	for _, m := range notes {
		err := h.sendNote(feathers, m)
		if err != nil {
			log.Error(err)
			errCount++
			continue
		}
	}
	if errCount > 0 {
		return errors.New("failed to send release notes")
	}
	return nil
}

func (h *Handler) sendNote(feathers *feather.Feathers, note models.ReleaseNote) error {
	// We should pool the addresses by contact type so that we only send one note per contact type
	addressPool := feathers.GetAddressPoolByTeamNames(note.TeamNames...)
	for contactType, addresses := range addressPool {
		err := h.Clients[contactType].Send(note.Content, feathers.Config.Messages.Subject, addresses)
		if err != nil {
			return errors.Wrapf(err, "failed to send note")
		}
		log.Infof("Release note successfully sent to %s via %s", strings.Join(addresses, ", "), contactType)
	}
	return nil
}

func (h *Handler) IsInitialised(contactType string) bool {
	_, ok := h.Clients[contactType]
	return ok
}
