package server

import (
	"github.com/gin-gonic/gin"
	"github.com/spring-financial-group/peacock/pkg/config"
	"github.com/spring-financial-group/peacock/pkg/git/github"
	"github.com/spring-financial-group/peacock/pkg/handlers"
	"github.com/spring-financial-group/peacock/pkg/health"
	"github.com/spring-financial-group/peacock/pkg/logger"
	"github.com/spring-financial-group/peacock/pkg/webhook/handler"
	"github.com/spring-financial-group/peacock/pkg/webhook/usecase"
)

func inject(cfg *config.Config, sources *DataSources) (*gin.Engine, error) {
	// Setup router
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	publicGroup := router.Group("/")
	publicGroup.Use(logger.Middleware())
	infraGroup := router.Group("/")

	scmFactory := github.NewClientFactory(cfg.SCM.Token)

	messageHandlers := handlers.InitMessageHandlers(
		cfg.MessageHandlers.Slack.Token,
		cfg.MessageHandlers.Webhook.URL,
		cfg.MessageHandlers.Webhook.Token,
		cfg.MessageHandlers.Webhook.Secret,
	)

	webhookUC := webhookuc.NewUseCase(&cfg.SCM, scmFactory, messageHandlers)

	// Setup handlers
	webhookhandler.NewHandler(&cfg.SCM, publicGroup, webhookUC)
	infraGroup.GET("/health", health.ServeHealth)

	return router, nil
}
