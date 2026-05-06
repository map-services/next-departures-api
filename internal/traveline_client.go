package internal

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/rm-hull/next-departures-api/internal/models"
)

type TravelineClient interface {
	GetNextDepartures(atcoCode string) ([]models.NextDeparture, error)
}

type travelineClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewTravelineClient() TravelineClient {
	return &travelineClient{
		baseURL: "https://www.traveline.info/stops/",
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *travelineClient) GetNextDepartures(atcoCode string) ([]models.NextDeparture, error) {
	url := c.baseURL + atcoCode
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch traveline page: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("traveline returned unexpected status: %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse traveline page: %w", err)
	}

	results := []models.NextDeparture{}
	london, err := time.LoadLocation("Europe/London")
	if err != nil {
		london = time.UTC
	}
	now := time.Now().In(london)

	doc.Find(".departure-board__item").Each(func(i int, s *goquery.Selection) {
		lineName := strings.TrimSpace(s.Find(".single-visit__name").Text())
		destination := strings.TrimSpace(s.Find(".single-visit__description").Text())
		timeStr := strings.TrimSpace(s.Find(".single-visit__arrival-time__cell").Text())

		if lineName == "" || destination == "" || timeStr == "" {
			return
		}

		// Determine if live or scheduled from sr-only text
		srText := s.Find(".sr-only").First().Text()
		isLive := strings.Contains(srText, "Live")

		departureTime := c.parseTime(timeStr, now, london)
		if departureTime == nil {
			return
		}

		departure := models.NextDeparture{
			LineName:    lineName,
			Destination: destination,
		}

		if isLive {
			departure.ExpectedArrivalTime = departureTime
		} else {
			departure.AimedArrivalTime = departureTime
		}

		results = append(results, departure)
	})

	return results, nil
}

var (
	minsRegex  = regexp.MustCompile(`^(\d+)\s*mins?$`)
	clockRegex = regexp.MustCompile(`^(\d{1,2}):(\d{2})$`)
)

func (c *travelineClient) parseTime(timeStr string, now time.Time, loc *time.Location) *time.Time {
	timeStr = strings.ToLower(timeStr)

	// Case 1: "due" or "now"
	if timeStr == "due" || timeStr == "now" {
		t := now
		return &t
	}

	// Case 2: "5 mins"
	if matches := minsRegex.FindStringSubmatch(timeStr); len(matches) > 1 {
		mins, _ := strconv.Atoi(matches[1])
		t := now.Add(time.Duration(mins) * time.Minute)
		return &t
	}

	// Case 3: "22:24"
	if matches := clockRegex.FindStringSubmatch(timeStr); len(matches) > 2 {
		hour, _ := strconv.Atoi(matches[1])
		min, _ := strconv.Atoi(matches[2])

		t := time.Date(now.Year(), now.Month(), now.Day(), hour, min, 0, 0, loc)

		// If the time is in the past (e.g. it's 23:50 and time is 00:05), it's probably tomorrow
		if t.Before(now.Add(-1 * time.Hour)) {
			t = t.AddDate(0, 0, 1)
		}

		return &t
	}

	return nil
}
