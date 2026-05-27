package internal

import (
	"context"
	"github.com/map-services/next-departures-api/internal/models"
	"github.com/robfig/cron/v3"
	"log/slog"
)

const CRON_SCHEDULE_NAPTAN = "@every 19h"
const CRON_SCHEDULE_RATE_LIMIT_RESET = "0 0 * * *"

func StartCron(repo NaptanRepository, fbManager FallbackManager) (*cron.Cron, error) {

	slog.Info("Starting CRON jobs", 
			  "naptan_schedule", CRON_SCHEDULE_NAPTAN, 
			  "rate_limit_reset_schedule", CRON_SCHEDULE_RATE_LIMIT_RESET)

	c := cron.New()
	if _, err := c.AddFunc(CRON_SCHEDULE_NAPTAN, func() {
		err := TransientDownload(models.NAPTAN_CSV_URL, repo.ImportCSV(context.Background(), 0))
		if err != nil {
			slog.Error("Error importing download NaPTAN dataset", "error", err)
		}
	}); err != nil {
		return nil, err
	}

	if _, err := c.AddFunc(CRON_SCHEDULE_RATE_LIMIT_RESET, func() {
		slog.Info("Resetting SIRI rate limit fallback flag")
		fbManager.SetSiriRateLimited(false)
	}); err != nil {
		return nil, err
	}

	c.Start()
	return c, nil
}
