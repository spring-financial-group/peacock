package delivery

import (
	"github.com/gin-gonic/gin"
	"github.com/spring-financial-group/peacock/pkg/domain"
	"github.com/spring-financial-group/peacock/pkg/models"
	"strings"
	"time"
)

type ReleaseHandler struct {
	releaseUc domain.ReleaseUseCase
}

func NewHandler(group *gin.RouterGroup, releaseUc domain.ReleaseUseCase) {
	handler := ReleaseHandler{
		releaseUc: releaseUc,
	}

	group.GET("/:environment/after/:startTime", handler.GetReleasesAfterDate)
}

// GetReleasesAfterDate godoc
// @Summary Get releases after a specific date
// @Description Get releases after a specific date
// @Tags release
// @Accept json
// @Produce json
// @Param environment path string true "Environment"
// @Param startTime path string true "Start Time"
// @Param teams query string false "Teams"
// @Success 200
// @Router /releases/{environment}/after/{startTime} [get]
func (h *ReleaseHandler) GetReleasesAfterDate(c *gin.Context) {
	environment := c.Param("environment")
	startTimeParam := c.Param("startTime")
	startTime, err := time.Parse(time.RFC3339, startTimeParam)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid date format"})
		return
	}

	teamsParam := c.Query("teams")
	var teams []string
	if teamsParam == "" {
		teams = []string{}
	} else {
		teams = strings.Split(teamsParam, ",")
	}

	releases, err := h.releaseUc.GetReleases(c, environment, startTime, teams)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, models.GetReleasesResponse{
		Releases: releases,
	})
}
