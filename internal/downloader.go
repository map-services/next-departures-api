package internal

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"

	"net/url"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
)

func applyDateTemplate(uri string) string {
	now := time.Now()
	yyyy := now.Format("2006")
	mm := now.Format("01")
	dd := now.Format("02")

	r := strings.NewReplacer(
		"{{yyyy}}", yyyy,
		"{{mm}}", mm,
		"{{dd}}", dd,
	)
	return r.Replace(uri)
}

func isValidUrl(uri string) bool {
	u, err := url.Parse(uri)
	return err == nil && (u.Scheme == "http" || u.Scheme == "https")
}

func TransientDownload(uri string, handler func(tmpfile string, header http.Header) error) error {
	uri = applyDateTemplate(uri)
	if !isValidUrl(uri) {
		return handler(uri, http.Header{})
	}

	slog.Info("Retrieving", "uri", uri)
	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	client := &http.Client{Timeout: 5 * time.Minute}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to fetch from %s: %w", uri, err)
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			slog.Error("failed to close body", "error", err)
		}
	}()

	if resp.StatusCode > 299 {
		return fmt.Errorf("error response from %s: %s", uri, resp.Status)
	}

	tmp, err := os.CreateTemp("", "download-*")
	if err != nil {
		return err
	}
	tmpfile := tmp.Name()

	lastModified := resp.Header.Get("Last-Modified")
	if lastModified == "" {
		lastModified = "unknown"
	}
	slog.Info("Remote last modified", "last_modified", lastModified)

	filesize := "unknown size"
	if resp.ContentLength >= 0 {
		filesize = humanize.Bytes(uint64(resp.ContentLength))
	}
	slog.Info("Downloading content", "size", filesize, "to", tmpfile)

	defer func() {
		slog.Info("Removing temporary file", "path", tmpfile)
		if err := os.Remove(tmpfile); err != nil {
			slog.Error("failed to remove file", "path", tmpfile, "error", err)
		}
	}()

	if _, err := io.Copy(tmp, resp.Body); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("failed to copy response body: %w", err)
	}

	if err := tmp.Close(); err != nil {
		return fmt.Errorf("failed to close temporary file: %w", err)
	}
	return handler(tmpfile, resp.Header)
}
