package models

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"
)

const NAPTAN_CSV_URL = "https://naptan.api.dft.gov.uk/v1/access-nodes?dataFormat=csv"

type NaPTAN struct {
	ATCOCode                string     `json:"atco_code,omitempty"`
	NaptanCode              string     `json:"naptan_code,omitempty"`
	PlateCode               string     `json:"plate_code,omitempty"`
	CleardownCode           string     `json:"cleardown_code,omitempty"`
	CommonName              string     `json:"common_name,omitempty"`
	CommonNameLang          string     `json:"-"`
	ShortCommonName         string     `json:"short_common_name,omitempty"`
	ShortCommonNameLang     string     `json:"-"`
	Landmark                string     `json:"landmark,omitempty"`
	LandmarkLang            string     `json:"-"`
	Street                  string     `json:"street,omitempty"`
	StreetLang              string     `json:"-"`
	Crossing                string     `json:"crossing,omitempty"`
	CrossingLang            string     `json:"-"`
	Indicator               string     `json:"indicator,omitempty"`
	IndicatorLang           string     `json:"-"`
	Bearing                 string     `json:"bearing,omitempty"`
	NptgLocalityCode        string     `json:"nptg_locality_code,omitempty"`
	LocalityName            string     `json:"locality_name,omitempty"`
	ParentLocalityName      string     `json:"parent_locality_name,omitempty"`
	GrandParentLocalityName string     `json:"grand_parent_locality_name,omitempty"`
	Town                    string     `json:"town,omitempty"`
	TownLang                string     `json:"-"`
	Suburb                  string     `json:"suburb,omitempty"`
	SuburbLang              string     `json:"-"`
	LocalityCentre          *bool      `json:"locality_centre,omitempty"`
	GridType                string     `json:"grid_type,omitempty"`
	Easting                 *int       `json:"easting,omitempty"`
	Northing                *int       `json:"northing,omitempty"`
	Longitude               *float64   `json:"longitude,omitempty"`
	Latitude                *float64   `json:"latitude,omitempty"`
	StopType                string     `json:"stop_type,omitempty"`
	BusStopType             string     `json:"bus_stop_type,omitempty"`
	TimingStatus            string     `json:"timing_status,omitempty"`
	DefaultWaitTime         string     `json:"default_wait_time,omitempty"`
	Notes                   string     `json:"notes,omitempty"`
	NotesLang               string     `json:"-"`
	AdministrativeAreaCode  string     `json:"administrative_area_code,omitempty"`
	CreationDateTime        *time.Time `json:"creation_date_time,omitempty"`
	ModificationDateTime    *time.Time `json:"modification_date_time,omitempty"`
	RevisionNumber          *int       `json:"revision_number,omitempty"`
	Modification            string     `json:"modification,omitempty"`
	Status                  string     `json:"status,omitempty"`
}

func FromTuple(data []string, headerMap map[string]int) (*NaPTAN, error) {
	val := func(header string) string {
		if idx, ok := headerMap[header]; ok && idx < len(data) {
			return data[idx]
		}
		return ""
	}

	localityCentre, err := parseBool(val("LocalityCentre"))
	if err != nil {
		return nil, err
	}
	easting, err := parseInt(val("Easting"))
	if err != nil {
		return nil, err
	}
	northing, err := parseInt(val("Northing"))
	if err != nil {
		return nil, err
	}
	longitude, err := parseFloat(val("Longitude"))
	if err != nil {
		return nil, err
	}
	latitude, err := parseFloat(val("Latitude"))
	if err != nil {
		return nil, err
	}
	creationDateTime, err := parseTime(val("CreationDateTime"))
	if err != nil {
		return nil, err
	}
	modificationDateTime, err := parseTime(val("ModificationDateTime"))
	if err != nil {
		return nil, err
	}
	revisionNumber, err := parseInt(val("RevisionNumber"))
	if err != nil {
		return nil, err
	}

	return &NaPTAN{
		ATCOCode:                val("ATCOCode"),
		NaptanCode:              val("NaptanCode"),
		PlateCode:               val("PlateCode"),
		CleardownCode:           val("CleardownCode"),
		CommonName:              val("CommonName"),
		CommonNameLang:          val("CommonNameLang"),
		ShortCommonName:         val("ShortCommonName"),
		ShortCommonNameLang:     val("ShortCommonNameLang"),
		Landmark:                val("Landmark"),
		LandmarkLang:            val("LandmarkLang"),
		Street:                  val("Street"),
		StreetLang:              val("StreetLang"),
		Crossing:                val("Crossing"),
		CrossingLang:            val("CrossingLang"),
		Indicator:               val("Indicator"),
		IndicatorLang:           val("IndicatorLang"),
		Bearing:                 val("Bearing"),
		NptgLocalityCode:        val("NptgLocalityCode"),
		LocalityName:            val("LocalityName"),
		ParentLocalityName:      val("ParentLocalityName"),
		GrandParentLocalityName: val("GrandParentLocalityName"),
		Town:                    val("Town"),
		TownLang:                val("TownLang"),
		Suburb:                  val("Suburb"),
		SuburbLang:              val("SuburbLang"),
		LocalityCentre:          localityCentre,
		GridType:                val("GridType"),
		Easting:                 easting,
		Northing:                northing,
		Longitude:               longitude,
		Latitude:                latitude,
		StopType:                val("StopType"),
		BusStopType:             val("BusStopType"),
		TimingStatus:            val("TimingStatus"),
		DefaultWaitTime:         val("DefaultWaitTime"),
		Notes:                   val("Notes"),
		NotesLang:               val("NotesLang"),
		AdministrativeAreaCode:  val("AdministrativeAreaCode"),
		CreationDateTime:        creationDateTime,
		ModificationDateTime:    modificationDateTime,
		RevisionNumber:          revisionNumber,
		Modification:            val("Modification"),
		Status:                  val("Status"),
	}, nil
}

func (n *NaPTAN) ToTuple() []any {
	strOrNil := func(s string) any {
		if s == "" {
			return nil
		}
		return s
	}

	derefInt := func(i *int) any {
		if i == nil {
			return nil
		}
		return *i
	}

	derefBool := func(b *bool) any {
		if b == nil {
			return nil
		}
		return *b
	}

	derefFloat := func(f *float64) any {
		if f == nil {
			return nil
		}
		return *f
	}

	return []any{
		strOrNil(n.ATCOCode),
		strOrNil(n.NaptanCode),
		strOrNil(n.PlateCode),
		strOrNil(n.CleardownCode),
		strOrNil(n.CommonName),
		strOrNil(n.CommonNameLang),
		strOrNil(n.ShortCommonName),
		strOrNil(n.ShortCommonNameLang),
		strOrNil(n.Landmark),
		strOrNil(n.LandmarkLang),
		strOrNil(n.Street),
		strOrNil(n.StreetLang),
		strOrNil(n.Crossing),
		strOrNil(n.CrossingLang),
		strOrNil(n.Indicator),
		strOrNil(n.IndicatorLang),
		strOrNil(n.Bearing),
		strOrNil(n.NptgLocalityCode),
		strOrNil(n.LocalityName),
		strOrNil(n.ParentLocalityName),
		strOrNil(n.GrandParentLocalityName),
		strOrNil(n.Town),
		strOrNil(n.TownLang),
		strOrNil(n.Suburb),
		strOrNil(n.SuburbLang),
		derefBool(n.LocalityCentre),
		strOrNil(n.GridType),
		derefInt(n.Easting),
		derefInt(n.Northing),
		derefFloat(n.Longitude),
		derefFloat(n.Latitude),
		strOrNil(n.StopType),
		strOrNil(n.BusStopType),
		strOrNil(n.TimingStatus),
		strOrNil(n.DefaultWaitTime),
		strOrNil(n.Notes),
		strOrNil(n.NotesLang),
		strOrNil(n.AdministrativeAreaCode),
		n.CreationDateTime,
		n.ModificationDateTime,
		derefInt(n.RevisionNumber),
		strOrNil(n.Modification),
		strOrNil(n.Status),
	}
}

func parseInt(s string) (*int, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, nil
	}
	i, err := strconv.Atoi(s)
	if err != nil {
		return nil, fmt.Errorf("failed to parse int from '%s': %w", s, err)
	}
	return &i, nil
}

func parseBool(s string) (*bool, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, nil
	}
	b, err := strconv.ParseBool(s)
	if err != nil {
		return nil, fmt.Errorf("failed to parse bool from '%s': %w", s, err)
	}
	return &b, nil
}

func parseFloat(s string) (*float64, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, nil
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse float from '%s': %w", s, err)
	}
	return &f, nil
}

func parseTime(s string) (*time.Time, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, nil
	}
	// Try parsing in order of most specific to least specific formats
	var t time.Time
	var err error

	// First try RFC3339Nano (with milliseconds and timezone)
	t, err = time.Parse(time.RFC3339Nano, s)
	if err == nil {
		return &t, nil
	}

	// Then try RFC3339 (with timezone, no milliseconds)
	t, err = time.Parse(time.RFC3339, s)
	if err == nil {
		return &t, nil
	}

	// Finally try simple format (no timezone, no milliseconds)
	t, err = time.Parse("2006-01-02T15:04:05", s)
	if err == nil {
		return &t, nil
	}

	return nil, fmt.Errorf("failed to parse time from '%s' with any supported format: %w", s, err)
}

type StopType struct {
	Code        string `json:"code"`
	Description string `json:"description"`
	Category    string `json:"category"`
}

//go:embed refdata/stop-types.json
var stopTypesJSON []byte

var StopTypes []StopType

func init() {
	err := json.Unmarshal(stopTypesJSON, &StopTypes)
	if err != nil {
		slog.Error("failed to unmarshal stop types JSON", "error", err)
		os.Exit(1)
	}
}
