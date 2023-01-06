package server

import (
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/spring-financial-group/peacock/pkg/config"
	"github.com/spring-financial-group/peacock/pkg/health"
	"github.com/spring-financial-group/peacock/pkg/webhook"
)

func inject(cfg *config.Config, sources *DataSources) (*gin.Engine, error) {
	// Setup router
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	publicGroup := router.Group("/")
	publicGroup.Use(gin.Logger())

	// Setup handlers
	webhooks, err := webhook.NewHandler(&cfg.SCM)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create webhook handler")
	}

	publicGroup.POST("/webhooks", webhooks.HandleEvents)
	publicGroup.GET("/health", health.ServeHealth)

	return router, nil
}
