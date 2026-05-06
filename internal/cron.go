package internal

import (
	"log"

	"github.com/rm-hull/next-departures-api/internal/models"
	"github.com/robfig/cron/v3"
)

const CRON_SCHEDULE_NAPTAN = "@every 19h"
const CRON_SCHEDULE_MIDNIGHT = "0 0 * * *"

func StartCron(repo NaptanRepository, fbManager FallbackManager) (*cron.Cron, error) {

	log.Printf("Starting CRON jobs: NaPTAN updates (schedule: %s), Fallback reset (schedule: %s)", CRON_SCHEDULE_NAPTAN, CRON_SCHEDULE_MIDNIGHT)

	c := cron.New()
	if _, err := c.AddFunc(CRON_SCHEDULE_NAPTAN, func() {
		err := TransientDownload(models.NAPTAN_CSV_URL, repo.ImportCSV(0))
		if err != nil {
			log.Printf("Error importing download NaPTAN dataset: %v", err)
		}
	}); err != nil {
		return nil, err
	}

	if _, err := c.AddFunc(CRON_SCHEDULE_MIDNIGHT, func() {
		log.Println("Resetting SIRI rate limit fallback flag")
		fbManager.SetSiriRateLimited(false)
	}); err != nil {
		return nil, err
	}

	c.Start()
	return c, nil
}
