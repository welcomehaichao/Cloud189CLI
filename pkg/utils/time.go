package utils

import (
	"time"
)

func ParseTime(timeStr string) (time.Time, error) {
	formats := []string{
		"2006-01-02 15:04:05 -07",
		"Jan 2, 2006 15:04:05 PM -07",
	}

	for _, format := range formats {
		if t, err := time.ParseInLocation(format, timeStr+" +08", time.Local); err == nil {
			return t, nil
		}
	}

	return time.Time{}, nil
}

func FormatTime(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}

func HTTPTime() string {
	return time.Now().UTC().Format(time.RFC1123)
}

func Timestamp() int64 {
	return time.Now().UnixNano() / 1e6
}

func TimestampSeconds() int64 {
	return time.Now().Unix()
}

func FormatDate(t time.Time) string {
	return t.Format("2006-01-0215:04:05.000")
}
