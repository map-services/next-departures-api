package internal

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestParseTime(t *testing.T) {
	london, _ := time.LoadLocation("Europe/London")
	now := time.Date(2026, 5, 6, 21, 0, 0, 0, london)
	client := &travelineClient{}

	tests := []struct {
		input    string
		expected time.Time
	}{
		{
			input:    "due",
			expected: now,
		},
		{
			input:    "now",
			expected: now,
		},
		{
			input:    "5 mins",
			expected: now.Add(5 * time.Minute),
		},
		{
			input:    "1 min",
			expected: now.Add(1 * time.Minute),
		},
		{
			input:    "22:24",
			expected: time.Date(2026, 5, 6, 22, 24, 0, 0, london),
		},
		{
			input:    "00:05",
			expected: time.Date(2026, 5, 7, 0, 5, 0, 0, london), // Tomorrow
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			actual := client.parseTime(tt.input, now, london)
			assert.NotNil(t, actual)
			assert.True(t, tt.expected.Equal(*actual), "expected %v, got %v", tt.expected, *actual)
		})
	}
}
