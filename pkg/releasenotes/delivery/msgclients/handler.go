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
	clients map[string]domain.MessageClient
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
		clients: clients,
	}
}

func (h *Handler) SendMessages(feathers *feather.Feathers, messages []models.ReleaseNote) error {
	var errCount int
	for _, m := range messages {
		err := h.sendMessage(feathers, m)
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

func (h *Handler) sendMessage(feathers *feather.Feathers, message models.ReleaseNote) error {
	// We should pool the addresses by contact type so that we only send one message per contact type
	addressPool := feathers.GetAddressPoolByTeamNames(message.TeamNames...)
	for contactType, addresses := range addressPool {
		err := h.clients[contactType].Send(message.Content, feathers.Config.Messages.Subject, addresses)
		if err != nil {
			return errors.Wrapf(err, "failed to send message")
		}
		log.Infof("ReleaseNote successfully sent to %s via %s", strings.Join(addresses, ", "), contactType)
	}
	return nil
}

func (h *Handler) IsInitialised(contactType string) bool {
	_, ok := h.clients[contactType]
	return ok
}
