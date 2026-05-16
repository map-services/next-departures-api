package internal

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"net/http"
	"time"

	"github.com/map-services/next-departures-api/internal/metrics"
	"github.com/map-services/next-departures-api/internal/models/siri"
	"github.com/prometheus/client_golang/prometheus"
)

type SiriClient interface {
	GetStopMonitoring(monitoringRef string) (*siri.Siri, int, error)
}

type siriClient struct {
	appId             string
	appKey            string
	endpoint          string
	previewInterval   string
	maximumStopVisits int
	httpClient        *http.Client
	metrics           *metrics.SiriMetrics
}

func NewSiriClient(appId, appKey string, reg prometheus.Registerer) SiriClient {
	return &siriClient{
		appId:             appId,
		appKey:            appKey,
		endpoint:          "https://transportapi.com/nextbuses",
		previewInterval:   "PT120M",
		maximumStopVisits: 10,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		metrics: metrics.NewSiriMetrics(reg),
	}
}

func (c *siriClient) GetStopMonitoring(monitoringRef string) (*siri.Siri, int, error) {
	req := siri.StopMonitoringRequest{
		Version: "1.0",
		Xmlns:   "http://www.siri.org.uk/siri",
		ServiceRequest: siri.ServiceRequest{
			RequestorRef: c.appId,
			StopMonitoringRequest: siri.StopMonitoringReq{
				Version:           "1.0",
				MonitoringRef:     monitoringRef,
				PreviewInterval:   c.previewInterval,
				MaximumStopVisits: c.maximumStopVisits,
			},
		},
	}

	body, err := xml.Marshal(req)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to marshal XML request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", c.endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, 0, fmt.Errorf("failed to create HTTP request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/xml")
	httpReq.SetBasicAuth(c.appId, c.appKey)

	start := time.Now()
	resp, err := c.httpClient.Do(httpReq)
	c.metrics.RecordHttpCall(start, "GetStopMonitoring", resp, err)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to send HTTP request: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	var siriResp siri.Siri
	err = xml.NewDecoder(resp.Body).Decode(&siriResp)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to decode XML response: %w", err)
	}

	count := 0
	for _, delivery := range siriResp.ServiceDelivery.StopMonitoringDelivery {
		count += len(delivery.MonitoredStopVisit)
	}
	c.metrics.RecordFetchedItems("GetStopMonitoring", count)

	return &siriResp, resp.StatusCode, nil
}
