package internal

import (
	"database/sql"
	_ "embed"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/kofalt/go-memoize"
	"github.com/map-services/next-departures-api/internal/metrics"
	"github.com/map-services/next-departures-api/internal/models"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/tavsec/gin-healthcheck/checks"
)

//go:embed sql/insert_naptan.sql
var insertNaptanSQL string

//go:embed sql/search_naptan.sql
var searchNaptanSQL string

//go:embed sql/last_updated.sql
var lastUpdatedSQL string

type NaptanRepository interface {
	ImportCSV(updateInterval int) func(string, http.Header) error
	Search(boundingBox []float64) ([]models.SearchResult, error)
	LastUpdated() (*time.Time, error)
	Close() error
	Check() checks.Check
}

type sqliteRepository struct {
	db      *sql.DB
	cache   *memoize.Memoizer
	metrics *metrics.SqlMetrics
}

func NewNaptanRepository(db *sql.DB) NaptanRepository {
	return &sqliteRepository{
		db:      db,
		cache:   memoize.NewMemoizer(60*time.Minute, 10*time.Minute),
		metrics: metrics.NewSqlMetrics(prometheus.DefaultRegisterer),
	}
}

func (repo *sqliteRepository) Close() error {
	return repo.db.Close()
}

func (repo *sqliteRepository) Check() checks.Check {
	return checks.SqlCheck{Sql: repo.db}
}

func (repo *sqliteRepository) LastUpdated() (*time.Time, error) {
	result, err, _ := memoize.Call(repo.cache, "last_updated", func() (*time.Time, error) {

		defer repo.metrics.Record(time.Now(), "lastUpdated")

		var lastUpdated sql.NullString
		err := repo.db.QueryRow(lastUpdatedSQL).Scan(&lastUpdated)
		if err != nil {
			return nil, fmt.Errorf("failed to execute last updated query: %w", err)
		}

		if !lastUpdated.Valid || strings.TrimSpace(lastUpdated.String) == "" {
			return nil, nil
		}

		parsed, err := parseTimeString(lastUpdated.String)
		if err != nil {
			return nil, fmt.Errorf("failed to parse last updated time: %w", err)
		}

		return parsed, nil
	})
	return result, err
}

func parseTimeString(value string) (*time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, nil
	}

	formats := []string{
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
		"2006-01-02 15:04:05Z07:00",
	}

	for _, format := range formats {
		parsed, err := time.Parse(format, value)
		if err == nil {
			return &parsed, nil
		}
	}

	return nil, fmt.Errorf("unsupported time format: %s", value)
}

func (repo *sqliteRepository) ImportCSV(updateInterval int) func(string, http.Header) error {
	return func(tmpfile string, header http.Header) error {
		defer repo.metrics.Record(time.Now(), "importCSV")

		tx, err := repo.db.Begin()
		if err != nil {
			return fmt.Errorf("failed to begin transaction: %w", err)
		}
		defer func() {
			if err != nil {
				if rbErr := tx.Rollback(); rbErr != nil {
					log.Printf("error rolling back transaction: %v", rbErr)
				}
			}
		}()

		stmt, err := tx.Prepare(insertNaptanSQL)
		if err != nil {
			return fmt.Errorf("failed to prepare statement: %w", err)
		}
		defer func() {
			if err := stmt.Close(); err != nil {
				log.Printf("failed to close statement: %v", err)
			}
		}()

		reader, err := os.Open(tmpfile)
		if err != nil {
			return fmt.Errorf("error opening file: %v", err)
		}
		defer func() {
			if err := reader.Close(); err != nil {
				log.Printf("failed to close file reader: %v", err)
			}
		}()

		count := 0
		for result := range ParseCSV(reader, true, models.FromTuple) {
			if result.Error != nil {
				log.Printf("error parsing CSV record: %v", result.Error)
				continue
			}
			count++
			if updateInterval > 0 && count%updateInterval == 0 {
				log.Printf("Processed %d records", result.LineNum)
			}

			_, err = stmt.Exec(result.Value.ToTuple()...)
			if err != nil {
				return fmt.Errorf("failed to execute individual insert: %w", err)
			}
		}
		log.Printf("Completed import: %d records processed", count)

		if err = tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit transaction: %w", err)
		}

		return nil
	}
}

func (repo *sqliteRepository) Search(boundingBox []float64) ([]models.SearchResult, error) {
	defer repo.metrics.Record(time.Now(), "search")

	if len(boundingBox) != 4 {
		return nil, fmt.Errorf("boundingBox must have 4 elements: min_lat, max_lat, min_lng, max_lng")
	}

	rows, err := repo.db.Query(searchNaptanSQL, boundingBox[1], boundingBox[3], boundingBox[0], boundingBox[2])
	if err != nil {
		return nil, fmt.Errorf("failed to execute search query: %w", err)
	}
	defer func() {
		_ = rows.Close()
	}()

	var results []models.SearchResult
	for rows.Next() {
		var rec models.NaPTAN

		var atcoCode sql.NullString
		var naptanCode sql.NullString
		var plateCode sql.NullString
		var cleardownCode sql.NullString
		var commonName sql.NullString
		var shortCommonName sql.NullString
		var landmark sql.NullString
		var street sql.NullString
		var crossing sql.NullString
		var indicator sql.NullString
		var bearing sql.NullString
		var nptgLocalityCode sql.NullString
		var localityName sql.NullString
		var parentLocalityName sql.NullString
		var grandParentLocalityName sql.NullString
		var town sql.NullString
		var suburb sql.NullString
		var gridType sql.NullString
		var stopType sql.NullString
		var busStopType sql.NullString
		var timingStatus sql.NullString
		var defaultWaitTime sql.NullString
		var notes sql.NullString
		var administrativeAreaCode sql.NullString
		var modification sql.NullString
		var status sql.NullString

		var localityCentre sql.NullBool
		var easting sql.NullInt64
		var northing sql.NullInt64
		var longitude sql.NullFloat64
		var latitude sql.NullFloat64
		var creationDateTime sql.NullTime
		var modificationDateTime sql.NullTime
		var revisionNumber sql.NullInt64

		err := rows.Scan(
			&atcoCode,
			&naptanCode,
			&plateCode,
			&cleardownCode,
			&commonName,
			&shortCommonName,
			&landmark,
			&street,
			&crossing,
			&indicator,
			&bearing,
			&nptgLocalityCode,
			&localityName,
			&parentLocalityName,
			&grandParentLocalityName,
			&town,
			&suburb,
			&localityCentre,
			&gridType,
			&easting,
			&northing,
			&longitude,
			&latitude,
			&stopType,
			&busStopType,
			&timingStatus,
			&defaultWaitTime,
			&notes,
			&administrativeAreaCode,
			&creationDateTime,
			&modificationDateTime,
			&revisionNumber,
			&modification,
			&status,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		rec.ATCOCode = atcoCode.String
		rec.NaptanCode = naptanCode.String
		rec.PlateCode = plateCode.String
		rec.CleardownCode = cleardownCode.String
		rec.CommonName = commonName.String
		rec.ShortCommonName = shortCommonName.String
		rec.Landmark = landmark.String
		rec.Street = street.String
		rec.Crossing = crossing.String
		rec.Indicator = indicator.String
		rec.Bearing = bearing.String
		rec.NptgLocalityCode = nptgLocalityCode.String
		rec.LocalityName = localityName.String
		rec.ParentLocalityName = parentLocalityName.String
		rec.GrandParentLocalityName = grandParentLocalityName.String
		rec.Town = town.String
		rec.Suburb = suburb.String
		rec.GridType = gridType.String
		rec.StopType = stopType.String
		rec.BusStopType = busStopType.String
		rec.TimingStatus = timingStatus.String
		rec.DefaultWaitTime = defaultWaitTime.String
		rec.Notes = notes.String
		rec.AdministrativeAreaCode = administrativeAreaCode.String
		rec.Modification = modification.String
		rec.Status = status.String

		if localityCentre.Valid {
			rec.LocalityCentre = &localityCentre.Bool
		}
		if easting.Valid {
			e := int(easting.Int64)
			rec.Easting = &e
		}
		if northing.Valid {
			val := int(northing.Int64)
			rec.Northing = &val
		}
		if longitude.Valid {
			rec.Longitude = &longitude.Float64
		}
		if latitude.Valid {
			rec.Latitude = &latitude.Float64
		}
		if creationDateTime.Valid {
			rec.CreationDateTime = &creationDateTime.Time
		}
		if modificationDateTime.Valid {
			rec.ModificationDateTime = &modificationDateTime.Time
		}
		if revisionNumber.Valid {
			r := int(revisionNumber.Int64)
			rec.RevisionNumber = &r
		}

		results = append(results, models.SearchResult{NaPTAN: rec})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return results, nil
}
