package cmd

import (
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/earthboundkid/versioninfo/v2"
	"github.com/getsentry/sentry-go"
	"github.com/joho/godotenv"
	"github.com/rm-hull/godx"

	"github.com/map-services/next-departures-api/internal"
)

// bootstrap initialises shared resources used by both the API server and import
// commands. It returns the repository, and an error if something failed during startup.
func bootstrap(dbPath string, debug bool) (internal.NaptanRepository, error) {
	if err := godotenv.Load(); err != nil {
		slog.Info("No .env file found")
	}

	environment := "development"
	if os.Getenv("ENVIRONMENT") != "" {
		environment = os.Getenv("ENVIRONMENT")
	}
	err := sentry.Init(sentry.ClientOptions{
		Dsn:         os.Getenv("SENTRY_DSN"),
		Debug:       debug,
		Release:     versioninfo.Revision[:7],
		Environment: environment,
		EnableLogs:  true,
	})
	if err != nil {
		return nil, fmt.Errorf("sentry initialization failed: %w", err)
	}
	defer sentry.Flush(2 * time.Second)

	logger := setupLogger("info")
	godx.Diagnostics(logger)

	db, err := internal.Connect(dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	if err := internal.Migrate("migrations", dbPath); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("failed to migrate SQL: %w", err)
	}

	repo := internal.NewNaptanRepository(db)

	return repo, nil
}

func setupLogger(level string) *slog.Logger {
	var programLevel = new(slog.LevelVar)
	switch level {
	case "debug":
		programLevel.Set(slog.LevelDebug)
	case "warn":
		programLevel.Set(slog.LevelWarn)
	case "error":
		programLevel.Set(slog.LevelError)
	default:
		programLevel.Set(slog.LevelInfo)
	}

	handler := slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
		Level: programLevel,
	})
	logger := slog.New(handler)
	slog.SetDefault(logger)
	return logger
}
