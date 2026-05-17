package internal

import (
	"context"
	"database/sql"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func TestLastUpdated_ParsesStringAggregate(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:?_loc=UTC&_datetime_format=rfc3339")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()

	_, err = db.Exec(`CREATE TABLE naptan (modification_date_time TIMESTAMP);`)
	if err != nil {
		t.Fatal(err)
	}

	const expectedValue = "2026-04-27T23:50:48Z"
	_, err = db.Exec(`INSERT INTO naptan (modification_date_time) VALUES (?);`, expectedValue)
	if err != nil {
		t.Fatal(err)
	}

	repo := NewNaptanRepository(db)
	lastUpdated, err := repo.LastUpdated(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if lastUpdated == nil {
		t.Fatal("expected lastUpdated to be non-nil")
	}

	expectedTime, err := time.Parse(time.RFC3339, expectedValue)
	if err != nil {
		t.Fatal(err)
	}

	if !lastUpdated.Equal(expectedTime) {
		t.Fatalf("expected %v, got %v", expectedTime, lastUpdated)
	}
}

func TestLastUpdated_ParsesSpaceSeparatedOffsetTimestamp(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:?_loc=UTC&_datetime_format=rfc3339")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()

	_, err = db.Exec(`CREATE TABLE naptan (modification_date_time TIMESTAMP);`)
	if err != nil {
		t.Fatal(err)
	}

	const expectedValue = "2026-04-27 21:39:12+00:00"
	_, err = db.Exec(`INSERT INTO naptan (modification_date_time) VALUES (?);`, expectedValue)
	if err != nil {
		t.Fatal(err)
	}

	repo := NewNaptanRepository(db)
	lastUpdated, err := repo.LastUpdated(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if lastUpdated == nil {
		t.Fatal("expected lastUpdated to be non-nil")
	}

	expectedTime, err := time.Parse(time.RFC3339, "2026-04-27T21:39:12Z")
	if err != nil {
		t.Fatal(err)
	}

	if !lastUpdated.Equal(expectedTime) {
		t.Fatalf("expected %v, got %v", expectedTime, lastUpdated)
	}
}

func TestLastUpdated_ReturnsNilWhenNoRows(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:?_loc=UTC&_datetime_format=rfc3339")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()

	_, err = db.Exec(`CREATE TABLE naptan (modification_date_time TIMESTAMP);`)
	if err != nil {
		t.Fatal(err)
	}

	repo := NewNaptanRepository(db)
	lastUpdated, err := repo.LastUpdated(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if lastUpdated != nil {
		t.Fatalf("expected nil lastUpdated, got %v", lastUpdated)
	}
}
