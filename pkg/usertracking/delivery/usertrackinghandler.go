package delivery

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/spring-financial-group/peacock/pkg/domain"
	"github.com/spring-financial-group/peacock/pkg/models"
)

type UserTrackingHandler struct {
	userTrackingUc domain.UserReleaseTrackingUseCase
}

func NewHandler(group *gin.RouterGroup, userTrackingUc domain.UserReleaseTrackingUseCase) {
	handler := UserTrackingHandler{
		userTrackingUc: userTrackingUc,
	}

	group.GET("/:userId/unviewed", handler.GetUnviewedReleases)
	group.POST("/:userId/mark-viewed", handler.MarkReleasesViewed)
	group.GET("/:userId/status", handler.GetUserStatus)
}

// GetUnviewedReleases godoc
// @Summary Get unviewed releases for a user
// @Description Get releases that a user has not viewed yet
// @Tags user-tracking
// @Accept json
// @Produce json
// @Param userId path string true "User ID"
// @Param environment query string true "Environment"
// @Success 200 {object} models.GetUnviewedReleasesResponse
// @Router /releases/{userId}/unviewed [get]
func (h *UserTrackingHandler) GetUnviewedReleases(c *gin.Context) {
	userID := c.Param("userId")
	environment := c.Query("environment")

	if environment == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "environment query parameter is required"})
		return
	}

	response, err := h.userTrackingUc.GetUnviewedReleases(c, userID, environment)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// MarkReleasesViewed godoc
// @Summary Mark releases as viewed by a user
// @Description Mark specific releases as viewed by a user
// @Tags user-tracking
// @Accept json
// @Produce json
// @Param userId path string true "User ID"
// @Param request body models.MarkViewedRequest true "Mark viewed request"
// @Success 200
// @Router /releases/{userId}/mark-viewed [post]
func (h *UserTrackingHandler) MarkReleasesViewed(c *gin.Context) {
	userID := c.Param("userId")

	var request models.MarkViewedRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.userTrackingUc.MarkReleasesViewed(c, userID, request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "releases marked as viewed"})
}

// GetUserStatus godoc
// @Summary Get user's release viewing status
// @Description Get the current viewing status for a user
// @Tags user-tracking
// @Accept json
// @Produce json
// @Param userId path string true "User ID"
// @Param environment query string true "Environment"
// @Success 200 {object} models.GetUserStatusResponse
// @Router /releases/{userId}/status [get]
func (h *UserTrackingHandler) GetUserStatus(c *gin.Context) {
	userID := c.Param("userId")
	environment := c.Query("environment")

	if environment == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "environment query parameter is required"})
		return
	}

	response, err := h.userTrackingUc.GetUserStatus(c, userID, environment)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}
