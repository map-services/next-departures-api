package routes

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rm-hull/next-departures-api/internal/models"
	"github.com/rm-hull/next-departures-api/internal/models/siri"
	"github.com/stretchr/testify/assert"
)

type mockSiriClient struct {
	getStopMonitoringFn func(monitoringRef string) (*siri.Siri, int, error)
}

func (m *mockSiriClient) GetStopMonitoring(monitoringRef string) (*siri.Siri, int, error) {
	return m.getStopMonitoringFn(monitoringRef)
}

type mockTravelineClient struct {
	getNextDeparturesFn func(atcoCode string) ([]models.NextDeparture, error)
}

func (m *mockTravelineClient) GetNextDepartures(atcoCode string) ([]models.NextDeparture, error) {
	return m.getNextDeparturesFn(atcoCode)
}

type mockFallbackManager struct {
	limited bool
}

func (m *mockFallbackManager) IsSiriRateLimited() bool     { return m.limited }
func (m *mockFallbackManager) SetSiriRateLimited(l bool) { m.limited = l }

func TestNextDepartures_RateLimitedFallback(t *testing.T) {
	gin.SetMode(gin.TestMode)

	fbManager := &mockFallbackManager{limited: false}
	siriClient := &mockSiriClient{
		getStopMonitoringFn: func(monitoringRef string) (*siri.Siri, int, error) {
			return &siri.Siri{
				ServiceDelivery: siri.ServiceDelivery{
					ErrorCondition: &siri.ErrorCondition{
						AccessNotAllowedError: &siri.Error{
							ErrorText: "Usage limits are exceeded",
						},
					},
				},
			}, http.StatusForbidden, nil
		},
	}
	travelineClient := &mockTravelineClient{
		getNextDeparturesFn: func(atcoCode string) ([]models.NextDeparture, error) {
			return []models.NextDeparture{
				{LineName: "Fallback", Destination: "Traveline"},
			}, nil
		},
	}

	r := gin.New()
	r.GET("/v1/next-departures/:stopId", NextDepartures(siriClient, travelineClient, fbManager))

	req, _ := http.NewRequest(http.MethodGet, "/v1/next-departures/123", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"line_name":"Fallback"`)
	assert.True(t, fbManager.limited, "Fallback flag should be set")
}

func TestNextDepartures_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	aimedTime := time.Date(2026, 5, 1, 12, 0, 0, 0, time.UTC)
	expectedTime := time.Date(2026, 5, 1, 12, 5, 0, 0, time.UTC)

	fbManager := &mockFallbackManager{limited: false}
	siriClient := &mockSiriClient{
		getStopMonitoringFn: func(monitoringRef string) (*siri.Siri, int, error) {
			return &siri.Siri{
				ServiceDelivery: siri.ServiceDelivery{
					StopMonitoringDelivery: []siri.StopMonitoringDelivery{
						{
							MonitoredStopVisit: []siri.MonitoredStopVisit{
								{
									MonitoredVehicleJourney: siri.MonitoredVehicleJourney{
										PublishedLineName: "42",
										DirectionName:     "Galaxy",
										OperatorRef:       "MARVIN",
										MonitoredCall: siri.MonitoredCall{
											AimedArrivalTime:    &aimedTime,
											ExpectedArrivalTime: &expectedTime,
										},
									},
								},
							},
						},
					},
				},
			}, http.StatusOK, nil
		},
	}
	travelineClient := &mockTravelineClient{}

	r := gin.New()
	r.GET("/v1/next-departures/:stopId", NextDepartures(siriClient, travelineClient, fbManager))

	req, _ := http.NewRequest(http.MethodGet, "/v1/next-departures/123", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"line_name":"42"`)
}

func TestNextDepartures_NoDepartures(t *testing.T) {
	gin.SetMode(gin.TestMode)

	fbManager := &mockFallbackManager{limited: false}
	siriClient := &mockSiriClient{
		getStopMonitoringFn: func(monitoringRef string) (*siri.Siri, int, error) {
			return &siri.Siri{
				ServiceDelivery: siri.ServiceDelivery{
					StopMonitoringDelivery: []siri.StopMonitoringDelivery{},
				},
			}, http.StatusOK, nil
		},
	}
	travelineClient := &mockTravelineClient{}

	r := gin.New()
	r.GET("/v1/next-departures/:stopId", NextDepartures(siriClient, travelineClient, fbManager))

	req, _ := http.NewRequest(http.MethodGet, "/v1/next-departures/123", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"results":[]`)
}

func TestNextDepartures_AccessDenied(t *testing.T) {
	gin.SetMode(gin.TestMode)

	fbManager := &mockFallbackManager{limited: false}
	siriClient := &mockSiriClient{
		getStopMonitoringFn: func(monitoringRef string) (*siri.Siri, int, error) {
			return &siri.Siri{
				ServiceDelivery: siri.ServiceDelivery{
					ErrorCondition: &siri.ErrorCondition{
						AccessNotAllowedError: &siri.Error{
							ErrorText: "Invalid API Key",
						},
					},
				},
			}, http.StatusUnauthorized, nil
		},
	}
	travelineClient := &mockTravelineClient{}

	r := gin.New()
	r.GET("/v1/next-departures/:stopId", NextDepartures(siriClient, travelineClient, fbManager))

	req, _ := http.NewRequest(http.MethodGet, "/v1/next-departures/123", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestNextDepartures_ClientError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	fbManager := &mockFallbackManager{limited: false}
	siriClient := &mockSiriClient{
		getStopMonitoringFn: func(monitoringRef string) (*siri.Siri, int, error) {
			return nil, 0, errors.New("network failure")
		},
	}
	travelineClient := &mockTravelineClient{}

	r := gin.New()
	r.GET("/v1/next-departures/:stopId", NextDepartures(siriClient, travelineClient, fbManager))

	req, _ := http.NewRequest(http.MethodGet, "/v1/next-departures/123", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
