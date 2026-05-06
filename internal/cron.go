package internal

import (
	"log"

	"github.com/rm-hull/next-departures-api/internal/models"
	"github.com/robfig/cron/v3"
)

const CRON_SCHEDULE_NAPTAN = "@every 19h"

func StartCron(repo NaptanRepository) (*cron.Cron, error) {

	log.Printf("Starting CRON job to update NaPTAN datasets (schedule: %s)", CRON_SCHEDULE_NAPTAN)

	c := cron.New()
	if _, err := c.AddFunc(CRON_SCHEDULE_NAPTAN, func() {
		err := TransientDownload(models.NAPTAN_CSV_URL, repo.ImportCSV(0))
		if err != nil {
			log.Printf("Error importing download NaPTAN dataset: %v", err)
		}
	}); err != nil {
		return nil, err
	}

	c.Start()
	return c, nil
}
