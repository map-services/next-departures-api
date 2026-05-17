package routes

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/map-services/next-departures-api/internal"
	"github.com/map-services/next-departures-api/internal/models"
)

func NextDepartures(siri internal.SiriClient, traveline internal.TravelineClient, fbManager internal.FallbackManager) func(c *gin.Context) {
	return func(c *gin.Context) {
		stopId := c.Param("stopId")

		if fbManager.IsSiriRateLimited() {
			fetchFromTraveline(c, traveline, stopId)
			return
		}

		siriResp, statusCode, err := siri.GetStopMonitoring(stopId)
		if err != nil {
			slog.Error("error while fetching next departures from SIRI", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "An internal server error occurred"})
			return
		}

		switch statusCode {
		case http.StatusOK:
			results := make([]models.NextDeparture, 0)
			if len(siriResp.ServiceDelivery.StopMonitoringDelivery) == 0 {
				c.JSON(http.StatusOK, models.NextDepartureResponse{
					Results:     results,
					Attribution: internal.ATTRIBUTION,
				})
				return
			}
			for _, visit := range siriResp.ServiceDelivery.StopMonitoringDelivery[0].MonitoredStopVisit {
				results = append(results, models.NextDeparture{
					LineName:            visit.MonitoredVehicleJourney.PublishedLineName,
					Destination:         visit.MonitoredVehicleJourney.DirectionName,
					OperatorRef:         visit.MonitoredVehicleJourney.OperatorRef,
					AimedArrivalTime:    visit.MonitoredVehicleJourney.MonitoredCall.AimedArrivalTime,
					ExpectedArrivalTime: visit.MonitoredVehicleJourney.MonitoredCall.ExpectedArrivalTime,
				})
			}

			c.JSON(http.StatusOK, models.NextDepartureResponse{
				Results:     results,
				Attribution: internal.ATTRIBUTION,
			})

		case http.StatusForbidden, http.StatusUnauthorized, http.StatusTooManyRequests:
			errMsg := "Access denied"
			if siriResp.ServiceDelivery.ErrorCondition != nil {
				if siriResp.ServiceDelivery.ErrorCondition.AccessNotAllowedError != nil {
					errMsg = siriResp.ServiceDelivery.ErrorCondition.AccessNotAllowedError.ErrorText
				} else if siriResp.ServiceDelivery.ErrorCondition.OtherError != nil {
					errMsg = siriResp.ServiceDelivery.ErrorCondition.OtherError.ErrorText
				}
			}

			isRateLimit := statusCode == http.StatusTooManyRequests ||
				(statusCode == http.StatusForbidden && strings.Contains(errMsg, "Usage limits are exceeded"))

			if isRateLimit {
				slog.Info("SIRI rate limit exceeded, switching to Traveline fallback", "stopId", stopId)
				fbManager.SetSiriRateLimited(true)
				fetchFromTraveline(c, traveline, stopId)
				return
			}

			slog.Error("unexpected HTTP status code from SIRI API", "status", statusCode, "error", errMsg)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "An internal server error occurred"})

		default:
			slog.Error("unexpected HTTP status code from SIRI API", "status", statusCode)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "An internal server error occurred"})
		}
	}
}

func fetchFromTraveline(c *gin.Context, client internal.TravelineClient, stopId string) {
	results, err := client.GetNextDepartures(stopId)
	if err != nil {
		slog.Error("error while fetching next departures from Traveline", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "An internal server error occurred"})
		return
	}

	c.JSON(http.StatusOK, models.NextDepartureResponse{
		Results:     results,
		Attribution: internal.ATTRIBUTION,
	})
}
