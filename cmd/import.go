package cmd

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/map-services/next-departures-api/internal"
	"github.com/map-services/next-departures-api/internal/models"
)

func Import(dbPath string) error {

	repo, err := bootstrap(dbPath, true)
	if err != nil {
		return err
	}
	defer func() {
		if err := repo.Close(); err != nil {
			slog.Error("failed to close repository", "error", err)
		}
	}()

	err = internal.TransientDownload(models.NAPTAN_CSV_URL, repo.ImportCSV(context.Background(), 421))
	if err != nil {
		return fmt.Errorf("failed to download NaPTAN dataset: %w", err)
	}

	return nil
}
