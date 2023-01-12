package health

import (
	"github.com/gin-gonic/gin"
)

// ServeHealth godoc
// @Summary Gets the health of the service
// @Description get the health of the dependencies of the service
// @Tags Health
// @Produce json
// @Success 200
// @Router /health [get]
func ServeHealth(c *gin.Context) {
	// Todo: Add health checks for other dependencies
	resp := struct {
		Success          bool    `json:"success"`
		ErrorDescription *string `json:"errorDescription"`
	}{
		Success:          true,
		ErrorDescription: nil,
	}

	c.JSON(200, resp)
}
