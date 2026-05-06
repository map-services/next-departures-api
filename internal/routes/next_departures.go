package routes

import (
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rm-hull/next-departures-api/internal"
	"github.com/rm-hull/next-departures-api/internal/models"
)

func NextDepartures(client internal.SiriClient) func(c *gin.Context) {
	return func(c *gin.Context) {
		stopId := c.Param("stopId")
		siri, statusCode, err := client.GetStopMonitoring(stopId)
		if err != nil {
			log.Printf("error while fetching next departures: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "An internal server error occurred"})
			return
		}

		switch statusCode {
		case http.StatusOK:
			results := make([]models.NextDeparture, 0)
			if len(siri.ServiceDelivery.StopMonitoringDelivery) == 0 {
				c.JSON(http.StatusOK, models.NextDepartureResponse{
					Results:     results,
					Attribution: internal.ATTRIBUTION,
				})
				return
			}
			for _, visit := range siri.ServiceDelivery.StopMonitoringDelivery[0].MonitoredStopVisit {
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

		case http.StatusBadRequest:
			errMsg := "Bad request to SIRI API"
			if siri.ServiceDelivery.ErrorCondition != nil && siri.ServiceDelivery.ErrorCondition.OtherError != nil {
				errMsg = siri.ServiceDelivery.ErrorCondition.OtherError.ErrorText
			}
			c.JSON(statusCode, gin.H{"error": errMsg})

		case http.StatusForbidden, http.StatusUnauthorized:
			errMsg := "Access denied"
			if siri.ServiceDelivery.ErrorCondition != nil && siri.ServiceDelivery.ErrorCondition.AccessNotAllowedError != nil {
				errMsg = siri.ServiceDelivery.ErrorCondition.AccessNotAllowedError.ErrorText
			}

			if statusCode == http.StatusForbidden && strings.Contains(errMsg, "Usage limits are exceeded") {
				now := time.Now().UTC()
				midnight := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, time.UTC)
				retryAfter := int(midnight.Sub(now).Seconds())
				c.Header("Retry-After", strconv.Itoa(retryAfter))
				c.Header("X-RateLimit-Reset", midnight.Format(time.RFC3339))
				c.JSON(http.StatusTooManyRequests, gin.H{"error": "Rate limit exceeded. Please try again after midnight."})
				return
			}

			log.Printf("unexpected HTTP status code (%d) from SIRI API: %s", statusCode, errMsg)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "An internal server error occurred"})

		default:
			log.Printf("unexpected HTTP status code (%d) from SIRI API", statusCode)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "An internal server error occurred"})
		}
	}
}
