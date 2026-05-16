package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/map-services/next-departures-api/internal/models"
)

func StopTypes(c *gin.Context) {
	c.JSON(http.StatusOK, models.StopTypes)
}
