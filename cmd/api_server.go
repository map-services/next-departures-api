package cmd

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/Depado/ginprom"
	"github.com/aurowora/compress"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rm-hull/next-departures-api/internal"
	"github.com/rm-hull/next-departures-api/internal/routes"

	healthcheck "github.com/tavsec/gin-healthcheck"
	"github.com/tavsec/gin-healthcheck/checks"
	hc_config "github.com/tavsec/gin-healthcheck/config"
)

func ApiServer(dbPath string, port int, debug bool) error {

	repo, err := bootstrap(dbPath, debug)
	if err != nil {
		return err
	}
	defer func() {
		if err := repo.Close(); err != nil {
			log.Printf("failed to close repository: %v", err)
		}
	}()

	fbManager := internal.NewFallbackManager()
	scheduler, err := internal.StartCron(repo, fbManager)
	if err != nil {
		return fmt.Errorf("failed to start CRON jobs: %w", err)
	}
	defer scheduler.Stop()

	appId := os.Getenv("TRANSPORTAPI_APP_ID")
	appKey := os.Getenv("TRANSPORTAPI_APP_KEY")
	siriClient := internal.NewSiriClient(appId, appKey, prometheus.DefaultRegisterer)
	travelineClient := internal.NewTravelineClient()

	r := gin.New()

	prom := ginprom.New(
		ginprom.Engine(r),
		ginprom.Path("/metrics"),
		ginprom.Ignore("/healthz"),
	)

	r.Use(
		gin.Recovery(),
		gin.LoggerWithWriter(gin.DefaultWriter, "/healthz", "/metrics"),
		prom.Instrument(),
		compress.Compress(),
		cors.Default(),
	)

	if debug {
		log.Println("WARNING: pprof endpoints are enabled and exposed. Do not run with this flag in production.")
		pprof.Register(r)
	}

	err = healthcheck.New(r, hc_config.DefaultConfig(), []checks.Check{
		repo.Check(),
	})
	if err != nil {
		return fmt.Errorf("failed to initialize healthcheck: %v", err)
	}

	r.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Route not found"})
	})

	r.NoMethod(func(c *gin.Context) {
		c.JSON(http.StatusMethodNotAllowed, gin.H{"error": "Method not allowed"})
	})

	v1 := r.Group("/v1/next-departures")
	v1.GET("/search", routes.Search(repo))
	v1.GET("/:stopId", routes.NextDepartures(siriClient, travelineClient, fbManager))

	refdata := v1.Group("/refdata")
	refdata.GET("/stop-types", routes.StopTypes)

	addr := fmt.Sprintf(":%d", port)
	log.Printf("Starting HTTP API Server on port %d...", port)
	if err := r.Run(addr); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("HTTP API Server failed to start on port %d: %v", port, err)
	}

	return nil
}
