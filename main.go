package main

import (
	"log/slog"
	"os"

	"github.com/map-services/next-departures-api/cmd"

	"github.com/spf13/cobra"
)

func main() {
	var err error
	var dbPath string
	var port int
	var debug bool

	rootCmd := &cobra.Command{
		Use:  "next-departures",
		Long: `Next Departures API`,
	}

	apiServerCmd := &cobra.Command{
		Use:   "api-server [--db <path>] [--port <port>] [--debug]",
		Short: "Start HTTP API server",
		Run: func(_ *cobra.Command, _ []string) {
			if err = cmd.ApiServer(dbPath, port, debug); err != nil {
				slog.Error("API Server failed", "error", err)
				os.Exit(1)
			}
		},
		}

		importCmd := &cobra.Command{
		Use:   "import [--db <path>]",
		Short: "Perform one-off import of bus stops from the GOV.UK API",
		Run: func(_ *cobra.Command, _ []string) {
			if err := cmd.Import(dbPath); err != nil {
				slog.Error("Import failed", "error", err)
				os.Exit(1)
			}
		},
		}

	apiServerCmd.Flags().IntVar(&port, "port", 8080, "Port to run HTTP server on")
	apiServerCmd.Flags().BoolVar(&debug, "debug", false, "Enable debugging (pprof) - WARING: do not enable in production")

	rootCmd.AddCommand(apiServerCmd)
	rootCmd.AddCommand(importCmd)
	rootCmd.PersistentFlags().StringVar(&dbPath, "db", "./data/next_departures.db", "Path to next-departures SQLite database")

	if err = rootCmd.Execute(); err != nil {
		slog.Error("root command execution failed", "error", err)
		os.Exit(1)
	}
}
