package server

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/spring-financial-group/peacock/pkg/config"
	"github.com/spring-financial-group/peacock/pkg/feathers"
	"github.com/spring-financial-group/peacock/pkg/git/github"
	"github.com/spring-financial-group/peacock/pkg/health"
	"github.com/spring-financial-group/peacock/pkg/logger"
	releasehandler "github.com/spring-financial-group/peacock/pkg/release/delivery"
	releaserepo "github.com/spring-financial-group/peacock/pkg/release/repository/mongodb"
	releaseuc "github.com/spring-financial-group/peacock/pkg/release/usecase"
	"github.com/spring-financial-group/peacock/pkg/releasenotes/delivery/msgclients"
	releasenotesuc "github.com/spring-financial-group/peacock/pkg/releasenotes/usecase"
	"github.com/spring-financial-group/peacock/pkg/webhook/handler"
	"github.com/spring-financial-group/peacock/pkg/webhook/usecase"
	"github.com/swaggest/swgui/v3cdn"
)

func inject(cfg *config.Config, data *DataSources) (*gin.Engine, error) {
	// Setup router
	gin.SetMode(gin.ReleaseMode)

	if !cfg.Cors.AllowAllOrigins && len(cfg.Cors.AllowOrigins) == 0 {
		panic("CORS_ALLOW_ORIGINS or CORS_ALLOW_ALL_ORIGINS must be set")
	}

	corsCfg := cors.DefaultConfig()
	corsCfg.AllowOrigins = cfg.Cors.AllowOrigins
	corsCfg.AllowAllOrigins = cfg.Cors.AllowAllOrigins
	corsCfg.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	corsCfg.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization"}

	router := gin.New()
	router.Use(cors.New(corsCfg))

	publicGroup := router.Group("/")
	publicGroup.Use(logger.Middleware())
	infraGroup := router.Group("/")

	scmClient := github.NewClient(cfg.SCM.User, cfg.SCM.Token)

	msgHandler := msgclients.NewMessageHandler(&cfg.MessageHandlers)

	notesUC := releasenotesuc.NewUseCase(msgHandler)

	feathersUC := feathers.NewUseCase()

	releaseRepo := releaserepo.NewRepository(*data.MongoDBClient)
	releaseUC := releaseuc.NewUseCase(releaseRepo)

	webhookUC := webhookuc.NewUseCase(&cfg.SCM, scmClient, notesUC, feathersUC, releaseUC)

	// Setup handlers
	webhookhandler.NewHandler(&cfg.SCM, publicGroup, webhookUC)
	releasehandler.NewHandler(publicGroup.Group("releases"), releaseUC)
	infraGroup.GET("/health", health.ServeHealth)
	infraGroup.GET("/swagger/v1/swagger.json", func(c *gin.Context) { c.File("docs/swagger.json") })
	infraGroup.GET("/swagger/index.html", gin.WrapH(v3cdn.NewHandler("Peacock API", "/swagger/v1/swagger.json", "/")))

	return router, nil
}
