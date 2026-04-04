package utils

import (
	"strings"
	"testing"
	"time"
)

func TestParseTime(t *testing.T) {
	tests := []struct {
		name     string
		timeStr  string
		wantErr  bool
		checkVal bool
	}{
		{
			name:     "standard format",
			timeStr:  "2024-01-15 10:30:00",
			wantErr:  false,
			checkVal: true,
		},
		{
			name:     "alternative format",
			timeStr:  "Jan 15, 2024 10:30:00 PM",
			wantErr:  false,
			checkVal: true,
		},
		{
			name:    "empty string",
			timeStr: "",
			wantErr: false,
		},
		{
			name:    "invalid format",
			timeStr: "invalid time",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseTime(tt.timeStr)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseTime() error = %v, wantError %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && tt.checkVal && result.IsZero() {
				t.Error("ParseTime() returned zero time for valid input")
			}
		})
	}
}

func TestFormatTime(t *testing.T) {
	tests := []struct {
		name string
		time time.Time
	}{
		{
			name: "standard time",
			time: time.Date(2024, 1, 15, 10, 30, 0, 0, time.Local),
		},
		{
			name: "zero time",
			time: time.Time{},
		},
		{
			name: "future time",
			time: time.Date(2025, 12, 31, 23, 59, 59, 0, time.Local),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatTime(tt.time)

			if !tt.time.IsZero() {
				expected := tt.time.Format("2006-01-02 15:04:05")
				if result != expected {
					t.Errorf("FormatTime() = %s, want %s", result, expected)
				}
			}
		})
	}
}

func TestHTTPTime(t *testing.T) {
	result := HTTPTime()

	if result == "" {
		t.Error("HTTPTime() returned empty string")
	}

	if !strings.Contains(result, ",") {
		t.Error("HTTPTime() should contain comma in RFC1123 format")
	}
}

func TestTimestamp(t *testing.T) {
	ts := Timestamp()

	if ts <= 0 {
		t.Error("Timestamp() should return positive value")
	}

	now := time.Now().UnixNano() / 1e6
	diff := now - ts

	if diff < 0 {
		diff = -diff
	}

	if diff > 1000 {
		t.Errorf("Timestamp() differs from current time by %d ms", diff)
	}
}

func TestTimestampSeconds(t *testing.T) {
	ts := TimestampSeconds()

	if ts <= 0 {
		t.Error("TimestampSeconds() should return positive value")
	}

	now := time.Now().Unix()
	diff := now - ts

	if diff < 0 {
		diff = -diff
	}

	if diff > 1 {
		t.Errorf("TimestampSeconds() differs from current time by %d seconds", diff)
	}
}

func TestFormatDate(t *testing.T) {
	tests := []struct {
		name string
		time time.Time
	}{
		{
			name: "standard time",
			time: time.Date(2024, 1, 15, 10, 30, 45, 0, time.Local),
		},
		{
			name: "zero time",
			time: time.Time{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatDate(tt.time)

			if !tt.time.IsZero() {
				expected := tt.time.Format("2006-01-0215:04:05.000")
				if result != expected {
					t.Errorf("FormatDate() = %s, want %s", result, expected)
				}
			}
		})
	}
}

func TestParseTimeFormatPreservation(t *testing.T) {
	original := "2024-01-15 10:30:00"
	parsed, err := ParseTime(original)
	if err != nil {
		t.Fatalf("ParseTime() error = %v", err)
	}

	formatted := FormatTime(parsed)

	if !strings.HasPrefix(formatted, "2024-01-15") {
		t.Errorf("Date not preserved: got %s", formatted)
	}
}
