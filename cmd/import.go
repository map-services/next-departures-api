package cmd

import (
	"fmt"
	"log"

	"github.com/rm-hull/next-departures-api/internal"
	"github.com/rm-hull/next-departures-api/internal/models"
)

func Import(dbPath string) error {

	repo, err := bootstrap(dbPath, true)
	if err != nil {
		return err
	}
	defer func() {
		if err := repo.Close(); err != nil {
			log.Printf("failed to close repository: %v", err)
		}
	}()

	err = internal.TransientDownload(models.NAPTAN_CSV_URL, repo.ImportCSV(421))
	if err != nil {
		return fmt.Errorf("failed to download NaPTAN dataset: %w", err)
	}

	return nil
}
